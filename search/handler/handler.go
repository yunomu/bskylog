package handler

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/aws/aws-lambda-go/events"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Client interface {
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
}

type Handler struct {
	s3Client          S3Client
	searchIndexBucket string
	publishBucket     string
	tmpDir            string
	logger            *slog.Logger
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
	}
}

func (h *Handler) Handle(ctx context.Context, req *events.APIGatewayV2HTTPRequest) *events.APIGatewayV2HTTPResponse {
	did, ok := req.PathParameters["did"]
	if !ok {
		h.logger.Info("Response",
			"status", http.StatusNotFound,
			"reason", "Path parameter `did` is not found",
			"pathParameters", req.PathParameters,
		)
		return &events.APIGatewayV2HTTPResponse{
			StatusCode: http.StatusNotFound,
		}
	}

	query, ok := req.QueryStringParameters["q"]
	if !ok {
		h.logger.Info("Response",
			"status", http.StatusBadRequest,
			"reason", "Query parameter `q` is not found",
			"pathParameters", req.QueryStringParameters,
		)
		return &events.APIGatewayV2HTTPResponse{
			StatusCode: http.StatusBadRequest,
		}
	}

	// TODO: retrieve /tmp/{did}
	var _ = did

	// TODO: if /tmp/{did} is not found, then GetObject from searchIndexBucket and store /tmp/{did}

	var _ = query

	return &events.APIGatewayV2HTTPResponse{
		StatusCode: http.StatusOK,
	}
}
