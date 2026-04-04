package handler

import (
	"context"
	"log/slog"

	"github.com/bluesky-social/indigo/api/bsky"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Client interface {
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
}

type Handler struct {
	s3Client S3Client
	bucket   string
	logger   *slog.Logger
}

func NewHandler(s3Client S3Client, bucket string, logger *slog.Logger) *Handler {
	return &Handler{
		s3Client: s3Client,
		bucket:   bucket,
		logger:   logger,
	}
}

type Request struct {
	DID   string                     `json:"did"`
	Posts []*bsky.FeedDefs_FeedViewPost `json:"posts"`
}

func (h *Handler) Handle(ctx context.Context, req *Request) error {
	h.logger.Info("IndexFunction received request", "did", req.DID, "num_posts", len(req.Posts))
	// ここに実際の処理を実装
	return nil
}
