package handler

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/bluesky-social/indigo/api/bsky"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"sort"

	"github.com/yunomu/bskylog/lib/index"
)

type S3Client interface {
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
}

type Handler struct {
	s3Client S3Client
	bucket   string
	tmpDir   string
	logger   *slog.Logger
}

func NewHandler(s3Client S3Client, bucket string, tmpDir string, logger *slog.Logger) *Handler {
	return &Handler{
		s3Client: s3Client,
		bucket:   bucket,
		tmpDir:   tmpDir,
		logger:   logger,
	}
}

type Item struct {
	Key      string                      `json:"key"`
	Position int                         `json:"pos"`
	Post     *bsky.FeedDefs_FeedViewPost `json:"post"`
}

func SortItems(items []*Item) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].Key != items[j].Key {
			return items[i].Key > items[j].Key
		}
		return items[i].Position > items[j].Position
	})
}

type Request struct {
	DID   string  `json:"did"`
	Items []*Item `json:"items"`
}

func (h *Handler) storeDB(ctx context.Context, filePath string, did string) error {
	getObjectOutput, err := h.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(h.bucket),
		Key:    aws.String(did),
	})
	if err != nil {
		var noSuchKey *types.NoSuchKey
		if errors.As(err, &noSuchKey) {
			h.logger.Info("Object not found in S3, creating new database", "bucket", h.bucket, "key", did)
			return nil // If object not found, we don't need to copy anything.
		} else {
			h.logger.Error("s3Client.GetObject", "err", err, "bucket", h.bucket, "key", did)
			return err
		}
	} else {
		defer getObjectOutput.Body.Close()
	}

	file, err := os.Create(filePath)
	if err != nil {
		h.logger.Error("os.Create", "err", err, "filePath", filePath)
		return err
	}
	defer file.Close()

	if _, err := io.Copy(file, getObjectOutput.Body); err != nil {
		h.logger.Error("io.Copy", "err", err, "filePath", filePath)
		return err
	}

	h.logger.Info("Successfully saved object to temporary file", "filePath", filePath)

	return nil
}

func (h *Handler) putItems(
	ctx context.Context,
	filePath string,
	items []*Item,
) error {
	db, err := gorm.Open(sqlite.Open(filePath), &gorm.Config{})
	if err != nil {
		h.logger.Error("gorm.Open", "err", err, "filePath", filePath)
		return err
	}
	sqlDB, err := db.DB()
	if err != nil {
		h.logger.Error("db.DB", "err", err)
		return err
	}
	defer sqlDB.Close()

	gormDB := index.NewGorm(db, index.GormOptionLogger(h.logger))

	for _, item := range items {
		if err := gormDB.Put(ctx, item.Key, item.Position, item.Post); err != nil {
			h.logger.Error("gormDB.Put", "err", err, "item", item)
			return err
		}
	}

	h.logger.Info("Successfully put posts into SQLite DB", "filepath", filePath, "num_items", len(items))

	return nil
}

func (h *Handler) Handle(ctx context.Context, req *Request) error {
	h.logger.Info("IndexFunction received request", "did", req.DID, "num_items", len(req.Items))

	filePath := filepath.Join(h.tmpDir, req.DID)
	defer os.Remove(filePath)

	if err := h.storeDB(ctx, filePath, req.DID); err != nil {
		return err
	}

	if err := h.putItems(ctx, filePath, req.Items); err != nil {
		return err
	}

	file, err := os.Open(filePath)
	if err != nil {
		h.logger.Error("os.Open", "err", err, "filePath", filePath)
		return err
	}
	defer file.Close()

	if _, err := h.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(h.bucket),
		Key:         aws.String(req.DID),
		Body:        file,
		ContentType: aws.String("application/vnd.sqlite3"),
	}); err != nil {
		h.logger.Error("s3Client.PutObject", "err", err, "bucket", h.bucket, "key", req.DID)
		return err
	}

	h.logger.Info("Successfully put SQLite DB to S3", "did", req.DID, "bucket", h.bucket, "key", req.DID)

	return nil
}
