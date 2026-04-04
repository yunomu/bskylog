package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/aws/aws-lambda-go/lambda"

	"github.com/yunomu/bskylog/index/handler"
)

func main() {
	ctx := context.Background()

	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	h := handler.NewHandler(logger)

	lambda.StartWithContext(ctx, h.Handle)
}
