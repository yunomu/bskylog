package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"

	"github.com/yunomu/bskylog/crawler/handler"
	"github.com/yunomu/bskylog/lib/userdb"
)

type LambdaClient interface {
	Invoke(ctx context.Context, params *lambda.InvokeInput, optFns ...func(*lambda.Options)) (*lambda.InvokeOutput, error)
}

type Handler struct {
	userDB          userdb.DB
	lambdaClient    LambdaClient
	crawlerFunction string

	logger *slog.Logger
}

func NewHandler(
	userDB userdb.DB,
	lambdaClient LambdaClient,
	crawlerFunction string,
	logger *slog.Logger,
) *Handler {
	return &Handler{
		userDB:          userDB,
		lambdaClient:    lambdaClient,
		crawlerFunction: crawlerFunction,
		logger:          logger,
	}
}

func (h *Handler) Handle(ctx context.Context) {
	var users []*userdb.User
	if err := h.userDB.Scan(ctx, func(user *userdb.User) error {
		users = append(users, user)
		return nil
	}); err != nil {
		h.logger.Error("userdb.Scan",
			"err", err,
		)
		return
	}

	for _, user := range users {
		payload := &handler.Request{
			Handle:   user.Handle,
			Password: user.Password,
			TimeZone: user.TimeZone,
		}

		var buf bytes.Buffer
		enc := json.NewEncoder(&buf)
		if err := enc.Encode(payload); err != nil {
			h.logger.Error("Request encode error",
				"err", err,
				"request", payload,
			)
			continue
		}

		if _, err := h.lambdaClient.Invoke(ctx, &lambda.InvokeInput{
			FunctionName:   aws.String(h.crawlerFunction),
			InvocationType: types.InvocationTypeEvent,
			Payload:        buf.Bytes(),
		}); err != nil {
			h.logger.Error("lambda.Invoke",
				"err", err,
				"function", h.crawlerFunction,
				"payload", buf.String(),
			)
			continue
		}

		h.logger.Info("Invoked",
			"function", h.crawlerFunction,
			"payload", buf.String(),
		)
	}
}
