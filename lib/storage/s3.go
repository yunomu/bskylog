package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"log/slog"
	"strings"

	"golang.org/x/sync/errgroup"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/bluesky-social/indigo/api/bsky"
)

// S3Client is an interface for S3 operations required by this package.
type S3Client interface {
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	ListObjectsV2(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error)
}

type S3 struct {
	client S3Client
	bucket string

	prefix      *string
	parallelism int

	logger *slog.Logger
}

type S3Option func(*S3)

func S3OptionPrefix(prefix string) S3Option {
	return func(s *S3) {
		if prefix == "" {
			s.prefix = nil
		} else {
			s.prefix = aws.String(prefix)
		}
	}
}

func S3OptionLogger(logger *slog.Logger) S3Option {
	return func(s *S3) {
		if logger == nil {
			s.logger = slog.Default()
		} else {
			s.logger = logger
		}
	}
}

func NewS3(client S3Client, bucket string, opts ...S3Option) *S3 {
	s := &S3{
		client: client,
		bucket: bucket,

		parallelism: 2,
		logger:      slog.Default(), // Default logger
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (s *S3) Scan(ctx context.Context, f func(key string, position int, post *bsky.FeedDefs_FeedViewPost) error) error {

	g, ctx := errgroup.WithContext(ctx)

	keyCh := make(chan string, s.parallelism)
	g.Go(func() error {
		defer close(keyCh)

		paginator := s3.NewListObjectsV2Paginator(s.client, &s3.ListObjectsV2Input{
			Bucket: aws.String(s.bucket),
			Prefix: s.prefix,
		})

		for paginator.HasMorePages() {
			output, err := paginator.NextPage(ctx)
			if err != nil {
				s.logger.Error("failed to list S3 objects", "err", err)
				return err
			}

			for _, obj := range output.Contents {
				obj := obj // capture loop variable

				key := aws.ToString(obj.Key)
				if strings.HasSuffix(key, "/") || strings.HasSuffix(key, "/index") {
					continue
				}

				select {
				case keyCh <- key:
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		}

		return nil
	})

	type postData struct {
		Key   string
		Posts []*bsky.FeedDefs_FeedViewPost
	}
	postCh := make(chan postData, s.parallelism)
	for i := 0; i < s.parallelism; i++ {
		g.Go(func() error {
			for key := range keyCh {
				output, err := s.client.GetObject(ctx, &s3.GetObjectInput{
					Bucket: aws.String(s.bucket),
					Key:    aws.String(key),
				})
				if err != nil {
					s.logger.Error("failed to get S3 object", "bucket", s.bucket, "key", key, "err", err)
					return err
				}
				defer output.Body.Close()

				var posts []*bsky.FeedDefs_FeedViewPost
				scanner := bufio.NewScanner(output.Body)
				for scanner.Scan() {
					line := scanner.Bytes()
					var post bsky.FeedDefs_FeedViewPost
					if err := json.Unmarshal(line, &post); err != nil {
						s.logger.Error("failed to unmarshal JSON from S3 object", "bucket", s.bucket, "key", key, "line", string(line), "err", err)
						return err
					}
					posts = append(posts, &post)
				}
				if err := scanner.Err(); err != nil {
					s.logger.Error("failed to read S3 object", "bucket", s.bucket, "key", key, "err", err)
					return err
				}

				select {
				case postCh <- postData{Key: key, Posts: posts}:
				case <-ctx.Done():
					return ctx.Err()
				}
			}

			return nil
		})
	}

	go func() {
		g.Wait()
		close(postCh)
	}()

	var consumerErr error
loop:
	for data := range postCh {
		select {
		case <-ctx.Done():
			break loop
		default:
			for position, post := range data.Posts {
				if err := f(data.Key, position, post); err != nil {
					consumerErr = err
					break loop
				}
			}
		}
	}

	if err := g.Wait(); err != nil {
		return err
	}

	return consumerErr
}
