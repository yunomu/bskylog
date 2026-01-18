package handler

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/xrpc"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"

	"github.com/yunomu/bskylog/lib/consumer"
	"github.com/yunomu/bskylog/lib/crawlerdb"
	"github.com/yunomu/bskylog/lib/processor"
	"github.com/yunomu/bskylog/lib/scanner"
)

type CloudFrontClient interface {
	CreateInvalidation(ctx context.Context, params *cloudfront.CreateInvalidationInput, optFns ...func(*cloudfront.Options)) (*cloudfront.CreateInvalidationOutput, error)
}

type Handler struct {
	xrpcHost         string
	crawlerDB        crawlerdb.DB
	s3Client         consumer.S3Client
	bucket           string
	cloudfrontClient CloudFrontClient
	distribution     string

	logger *slog.Logger
}

func NewHandler(
	xrpcHost string,
	crawlerDB crawlerdb.DB,
	s3Client consumer.S3Client,
	bucket string,
	cloudfrontClient CloudFrontClient,
	distribution string,
	logger *slog.Logger,
) *Handler {
	return &Handler{
		xrpcHost:         xrpcHost,
		crawlerDB:        crawlerDB,
		s3Client:         s3Client,
		bucket:           bucket,
		cloudfrontClient: cloudfrontClient,
		distribution:     distribution,
		logger:           logger,
	}
}

type Request struct {
	Handle   string `json:"handle"`
	Password string `json:"password"`
	TimeZone int    `json:"timezone"`
}

func (h *Handler) Handle(ctx context.Context, req *Request) {
	if req.Handle == "" {
		h.logger.Error("handle is empty")
		return
	}
	if req.Password == "" {
		h.logger.Error("password is empty")
		return
	}

	loc := time.FixedZone(fmt.Sprintf("%dmin", req.TimeZone), req.TimeZone*60)

	xrpcClient := &xrpc.Client{
		Host: h.xrpcHost,
	}

	session, err := atproto.ServerCreateSession(ctx, xrpcClient, &atproto.ServerCreateSession_Input{
		Identifier: req.Handle,
		Password:   req.Password,
	})
	if err != nil {
		h.logger.Error("ServerCreateSession",
			"err", err,
			"identifier", req.Handle,
			"password", req.Password,
		)
		return
	}

	ts, err := h.crawlerDB.Get(ctx, session.Did)
	if err != nil {
		h.logger.Error("crawldb.Get",
			"err", err,
			"did", session.Did,
		)
		return
	}

	xrpcClient.Auth = &xrpc.AuthInfo{
		AccessJwt:  session.AccessJwt,
		RefreshJwt: session.RefreshJwt,
		Did:        session.Did,
		Handle:     session.Handle,
	}

	var first *consumer.TerminalValue
	var updatedKeys []string
	p := processor.New(
		scanner.NewXRPCScanner(
			xrpcClient,
			session.Did,
			scanner.SetLogger(h.logger.With("module", "scanner")),
		),
		consumer.NewDailyJSONRecordS3(
			h.s3Client,
			h.bucket,
			session.Did,
			loc,
			consumer.SetDailyJSONRecordS3Logger(h.logger.With("module", "consumer")),
			consumer.SetDailyJSONRecordS3TerminalValue(
				&consumer.TerminalValue{
					TimeStamp: ts.Timestamp,
					Cid:       ts.LatestCid,
				},
			),
			consumer.SetDailyJSONRecordS3FirstValueFunc(
				func(v *consumer.TerminalValue) {
					first = v
				},
			),
			consumer.SetDailyJSONRecordS3KeyUpdateFunc(
				func(key string) {
					updatedKeys = append(updatedKeys, "/"+key)
				},
			),
		),
	)

	if err := p.Proc(ctx); err != nil {
		h.logger.Error("Proc",
			"err", err,
		)
		return
	}

	if err := p.Close(ctx); err != nil {
		h.logger.Warn("Proc close",
			"err", err,
		)
		// continue
	}

	if len(updatedKeys) != 0 {
		if _, err := h.cloudfrontClient.CreateInvalidation(ctx, &cloudfront.CreateInvalidationInput{
			DistributionId: aws.String(h.distribution),
			InvalidationBatch: &types.InvalidationBatch{
				CallerReference: aws.String(strconv.FormatInt(time.Now().Unix(), 10)),
				Paths: &types.Paths{
					Quantity: aws.Int32(int32(len(updatedKeys))),
					Items:    updatedKeys,
				},
			},
		}); err != nil {
			h.logger.Error("cloudfront.CreateInvalidation",
				"err", err,
				"distributionId", h.distribution,
				"paths", updatedKeys,
			)
			return
		}
	}

	if err := h.crawlerDB.Put(ctx, &crawlerdb.Timestamp{
		Did:       session.Did,
		LatestCid: first.Cid,
		Timestamp: first.TimeStamp,
	}); err != nil {
		h.logger.Error("crawldb.Get",
			"err", err,
			"did", session.Did,
		)
		return
	}
}
