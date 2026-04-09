package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/yunomu/bskylog/search/handler"
)

func main() {
	ctx := context.Background()

	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		logger.Error("config.LoadDefaultConfig", "err", err)
		os.Exit(1)
	}

	searchIndexBucket := os.Getenv("SEARCH_INDEX_BUCKET")
	publishBucket := os.Getenv("PUBLISH_BUCKET")
	tmpDir := os.Getenv("TMP_DIR")
	logger.Info("Start",
		"searchIndexBucket", searchIndexBucket,
		"publishBucket", publishBucket,
		"tmpDir", tmpDir,
	)

	h := handler.NewHandler(
		s3.NewFromConfig(cfg),
		searchIndexBucket,
		publishBucket,
		handler.WithTmpDir(tmpDir),
		handler.WithLogger(logger.With("module", "handler")),
		handler.WithLimit(100),
	)

	lambda.StartWithContext(ctx, h.Handle)
}
