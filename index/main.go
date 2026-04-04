package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/yunomu/bskylog/index/handler"
)

func main() {
	ctx := context.Background()

	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		logger.Error("config.LoadDefaultConfig", "err", err)
		os.Exit(1)
	}

	s3Client := s3.NewFromConfig(cfg)

	bucket := os.Getenv("SEARCH_INDEX_BUCKET")
	if bucket == "" {
		logger.Error("SEARCH_INDEX_BUCKET is not set")
		os.Exit(1)
	}

	h := handler.NewHandler(s3Client, bucket, "/tmp", logger)

	lambda.StartWithContext(ctx, h.Handle)
}
