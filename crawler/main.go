package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/aws/aws-lambda-go/lambda"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/yunomu/bskylog/lib/crawlerdb"

	"github.com/yunomu/bskylog/crawler/handler"
)

var (
	logger *slog.Logger
	debug  bool
)

func init() {
	debug = os.Getenv("DEBUG") != ""

	levelVar := new(slog.LevelVar)
	if debug {
		levelVar.Set(slog.LevelDebug)
	}
	logger = slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: levelVar,
	}))
}

func main() {
	region := os.Getenv("REGION")
	bucket := os.Getenv("BUCKET")
	distribution := os.Getenv("DISTRIBUTION")
	crawlerTable := os.Getenv("CRAWLER_TABLE")
	bskyHost := os.Getenv("BSKY_HOST")

	logger.Info("Init",
		"region", region,
		"bucket", bucket,
		"distribution", distribution,
		"crawlerTable", crawlerTable,
		"bskyHost", bskyHost,
	)

	ctx := context.Background()

	awsCfg, err := config.LoadDefaultConfig(ctx,
		func(opt *config.LoadOptions) error {
			opt.Region = region
			return nil
		},
	)
	if err != nil {
		logger.Error("LoadDefaultConfig", "err", err)
		return
	}

	h := handler.NewHandler(
		bskyHost,
		crawlerdb.NewDynamoDB(
			dynamodb.NewFromConfig(awsCfg),
			crawlerTable,
		),
		s3.NewFromConfig(awsCfg),
		bucket,
		cloudfront.NewFromConfig(awsCfg),
		distribution,
		logger.With("module", "handler"),
	)

	lambda.Start(h.Handle)
}
