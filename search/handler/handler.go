package handler

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/glebarez/sqlite"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/bluesky-social/indigo/api/bsky"

	indexhandler "github.com/yunomu/bskylog/index/handler"
	"github.com/yunomu/bskylog/lib/index"
)

var ErrIndexNotPrepared = errors.New("index not prepared")

type S3Client interface {
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
}

type Handler struct {
	s3Client          S3Client
	searchIndexBucket string
	publishBucket     string
	tmpDir            string
	logger            *slog.Logger
	parallelism       int
}

func NewHandler(
	s3Client S3Client,
	searchIndexBucket string,
	publishBucket string,
	tmpDir string,
	logger *slog.Logger,
) *Handler {
	return &Handler{
		s3Client:          s3Client,
		searchIndexBucket: searchIndexBucket,
		publishBucket:     publishBucket,
		tmpDir:            tmpDir,
		logger:            logger,
		parallelism:       2,
	}
}

func (h *Handler) retrieveIndexFile(ctx context.Context, did string, filePath string) error {
	if _, err := os.Stat(filePath); err == nil {
		h.logger.Debug("Index file already exists in /tmp", "path", filePath)
		return nil
	} else if !os.IsNotExist(err) {
		h.logger.Error("Failed to check file existence in /tmp", "err", err, "path", filePath)
		return err
	}

	h.logger.Debug("Index file not found in /tmp, retrieving from S3", "did", did, "bucket", h.searchIndexBucket)
	output, err := h.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &h.searchIndexBucket,
		Key:    &did,
	})
	if err != nil {
		if _, ok := err.(*types.NoSuchKey); ok {
			h.logger.Info("Index file not found in S3", "did", did, "bucket", h.searchIndexBucket)
			return ErrIndexNotPrepared
		}
		h.logger.Error("Failed to get object from S3", "err", err, "did", did, "bucket", h.searchIndexBucket)
		return err
	}
	defer output.Body.Close()

	file, err := os.Create(filePath)
	if err != nil {
		h.logger.Error("Failed to create file in /tmp", "err", err, "path", filePath)
		return err
	}
	defer file.Close()

	if _, err := io.Copy(file, output.Body); err != nil {
		h.logger.Error("Failed to write S3 object to file", "err", err, "path", filePath)
		return err
	}

	h.logger.Info("Successfully retrieved and stored index file", "path", filePath)
	return nil
}

func extractUniqueKeys(searchResults []*index.SearchResult) map[string][]int {
	ret := make(map[string][]int)
	for _, r := range searchResults {
		ret[r.Key] = append(ret[r.Key], r.Position)
	}

	return ret
}

func (h *Handler) getPostsFromSearchResults(ctx context.Context, searchResults []*index.SearchResult) ([]*bsky.FeedDefs_FeedViewPost, error) {
	keyMap := extractUniqueKeys(searchResults)

	g, ctx := errgroup.WithContext(ctx)

	keyCh := make(chan string, h.parallelism)
	g.Go(func() error {
		defer close(keyCh)

		for key := range keyMap {
			select {
			case keyCh <- key:
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		return nil
	})

	itemCh := make(chan *indexhandler.Item, len(searchResults))
	for i := 0; i < h.parallelism; i++ {
		g.Go(func() error {
			for key := range keyCh {
				out, err := h.s3Client.GetObject(ctx, &s3.GetObjectInput{
					Bucket: &h.publishBucket,
					Key:    &key,
				})
				if err != nil {
					h.logger.Error("GetObject",
						"bucket", h.publishBucket,
						"key", key,
						"err", err,
					)
					return err
				}
				defer out.Body.Close()

				posMap := make(map[int]bool)
				for _, p := range keyMap[key] {
					posMap[p] = true
				}

				scanner := bufio.NewScanner(out.Body)
				idx := 0
				for scanner.Scan() {
					if !posMap[idx] {
						idx++
						continue
					}

					var post bsky.FeedDefs_FeedViewPost
					if err := json.Unmarshal(scanner.Bytes(), &post); err != nil {
						h.logger.Error("Failed to unmarshal JSON record", "err", err, "key", key, "line", idx)
						return err
					}

					select {
					case itemCh <- &indexhandler.Item{
						Post:     &post,
						Key:      key,
						Position: idx,
					}:
					case <-ctx.Done():
						return ctx.Err()
					}

					idx++
				}

				if err := scanner.Err(); err != nil {
					h.logger.Error("Failed to scan S3 object body", "err", err, "key", key)
					return err
				}
			}

			return nil
		})
	}

	go func() {
		defer close(itemCh)
		g.Wait()
	}()

	var items []*indexhandler.Item
	for item := range itemCh {
		items = append(items, item)
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	indexhandler.SortItems(items)

	var ret []*bsky.FeedDefs_FeedViewPost
	for _, item := range items {
		ret = append(ret, item.Post)
	}

	return ret, nil
}

func (h *Handler) Handle(ctx context.Context, req *events.LambdaFunctionURLRequest) (*events.LambdaFunctionURLResponse, error) {
	path := req.RawPath
	const prefix = "/search/"
	if !strings.HasPrefix(path, prefix) {
		h.logger.Info("Response",
			"status", http.StatusBadRequest,
			"reason", "Path does not start with /search/",
			"rawPath", req.RawPath,
		)
		return &events.LambdaFunctionURLResponse{
			StatusCode: http.StatusBadRequest,
		}, nil
	}
	did := strings.TrimPrefix(path, prefix)
	if did == "" {
		h.logger.Info("Response",
			"status", http.StatusBadRequest,
			"reason", "DID is empty in path",
			"rawPath", req.RawPath,
		)
		return &events.LambdaFunctionURLResponse{
			StatusCode: http.StatusBadRequest,
		}, nil
	}

	query, ok := req.QueryStringParameters["q"]
	if !ok {
		h.logger.Info("Response",
			"status", http.StatusBadRequest,
			"reason", "Query parameter `q` is not found",
			"queryStringParameters", req.QueryStringParameters,
		)
		return &events.LambdaFunctionURLResponse{
			StatusCode: http.StatusBadRequest,
		}, nil
	}

	filePath := filepath.Join(h.tmpDir, did)
	if err := h.retrieveIndexFile(ctx, did, filePath); err != nil {
		if errors.Is(err, ErrIndexNotPrepared) {
			h.logger.Info("Index not prepared for did", "did", did)
			return &events.LambdaFunctionURLResponse{
				StatusCode: http.StatusInternalServerError,
				Body:       ErrIndexNotPrepared.Error(),
			}, nil
		}
		h.logger.Error("Failed to retrieve index file", "err", err, "did", did)
		return &events.LambdaFunctionURLResponse{
			StatusCode: http.StatusInternalServerError,
		}, nil
	}

	db, err := gorm.Open(sqlite.Open(filePath), &gorm.Config{})
	if err != nil {
		h.logger.Error("Failed to open SQLite database", "err", err, "path", filePath)
		return &events.LambdaFunctionURLResponse{
			StatusCode: http.StatusInternalServerError,
		}, nil
	}

	idx := index.NewGorm(db, index.GormOptionLogger(h.logger))

	searchResults, err := idx.Search(ctx, query)
	if err != nil {
		h.logger.Error("Failed to perform search", "err", err, "query", query)
		return &events.LambdaFunctionURLResponse{
			StatusCode: http.StatusInternalServerError,
		}, nil
	}

	posts, err := h.getPostsFromSearchResults(ctx, searchResults)
	if err != nil {
		h.logger.Error("Failed to get posts from search results", "err", err)
		return &events.LambdaFunctionURLResponse{
			StatusCode: http.StatusInternalServerError,
		}, nil
	}

	jsonBytes, err := json.Marshal(posts)
	if err != nil {
		h.logger.Error("Failed to marshal posts to JSON", "err", err)
		return &events.LambdaFunctionURLResponse{
			StatusCode: http.StatusInternalServerError,
		}, nil
	}

	return &events.LambdaFunctionURLResponse{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(jsonBytes),
	}, nil
}
