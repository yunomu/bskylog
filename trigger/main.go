package main

import (
	"context"
	"log/slog"
	"os"

	lambdaserver "github.com/aws/aws-lambda-go/lambda"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/lambda"

	"github.com/yunomu/bskylog/lib/userdb"

	"github.com/yunomu/bskylog/trigger/handler"
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
	crawlerFunction := os.Getenv("CRAWLER_FUNCTION")
	userTable := os.Getenv("USER_TABLE")
	userHandleIndex := os.Getenv("USER_HANDLE_INDEX")

	logger.Info("Init",
		"region", region,
		"crawlerFunction", crawlerFunction,
		"userTable", userTable,
		"userHandleIndex", userHandleIndex,
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
		userdb.NewDynamoDB(
			dynamodb.NewFromConfig(awsCfg),
			userTable,
			userHandleIndex,
		),
		lambda.NewFromConfig(awsCfg),
		crawlerFunction,
		logger.With("module", "handler"),
	)

	lambdaserver.Start(h.Handle)
}
