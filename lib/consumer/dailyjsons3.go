package consumer

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	smithy "github.com/aws/smithy-go"

	"github.com/bluesky-social/indigo/api/bsky"
)

type S3Client interface {
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
}

type DailyJSONRecordS3 struct {
	s3Client S3Client
	bucket   string
	baseDir  string
	location *time.Location

	year  int
	month time.Month
	day   int

	key     string
	buf     *bytes.Buffer
	encoder *json.Encoder

	logger *slog.Logger
}

var _ Consumer = (*DailyJSONRecordS3)(nil)

type DailyJSONRecordS3Option func(c *DailyJSONRecordS3)

func SetDailyJSONRecordS3Logger(logger *slog.Logger) DailyJSONRecordS3Option {
	return func(c *DailyJSONRecordS3) {
		if logger == nil {
			c.logger = slog.Default()
		} else {
			c.logger = logger
		}
	}
}

func NewDailyJSONRecordS3(
	s3Client S3Client,
	bucket string,
	baseDir string,
	location *time.Location,
	opts ...DailyJSONRecordS3Option,
) *DailyJSONRecordS3 {
	ret := &DailyJSONRecordS3{
		s3Client: s3Client,
		bucket:   bucket,
		baseDir:  baseDir,
		location: location,
		logger:   slog.Default(),
	}
	for _, f := range opts {
		f(ret)
	}
	return ret
}

func (c *DailyJSONRecordS3) ensureStream(ctx context.Context, now time.Time) error {
	year, month, day := now.Date()
	if day == c.day && month == c.month && year == c.year {
		return nil
	}

	c.Close(ctx)

	c.year = year
	c.month = month
	c.day = day

	c.key = fmt.Sprintf("%d/%04d/%02d/%02d", c.baseDir, year, int(month), day)

	c.buf = &bytes.Buffer{}

	c.encoder = json.NewEncoder(c.buf)

	return nil
}

func (c *DailyJSONRecordS3) Consume(ctx context.Context, post *bsky.FeedDefs_FeedViewPost) error {
	record, ok := post.Post.Record.Val.(*bsky.FeedPost)
	if !ok {
		c.logger.Warn("record is not post type", "cid", post.Post.Cid)

		// skip
		return nil
	}

	t, err := time.Parse(time.RFC3339Nano, record.CreatedAt)
	if err != nil {
		c.logger.Warn("time parse error",
			"cid", post.Post.Cid,
			"created_at", record.CreatedAt,
			"err", err,
		)

		// skip
		return nil
	}
	t = t.In(c.location)

	if err := c.ensureStream(ctx, t); err != nil {
		c.logger.Error("ensureStream() error",
			"baseDir", c.baseDir,
			"time", t,
		)
		return err
	}

	if err := c.encoder.Encode(post); err != nil {
		return err
	}

	c.logger.Info("Consume", "time", record.CreatedAt, "cid", post.Post.Cid)
	return nil
}

func (c *DailyJSONRecordS3) appendObject(ctx context.Context) error {
	s3Out, err := c.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(c.key),
	})
	if err != nil {
		var opErr *smithy.OperationError
		var noSuchKey *types.NoSuchKey
		if errors.As(err, &opErr) && errors.As(opErr.Err, &noSuchKey) {
			// do nothing
			return nil
		} else {
			c.logger.Error("s3.GetObject",
				"bucket", c.bucket,
				"key", c.key,
			)
			return err
		}
	}
	defer s3Out.Body.Close()

	_, err = c.buf.ReadFrom(s3Out.Body)

	return err
}

func (c *DailyJSONRecordS3) Close(ctx context.Context) {
	if c.buf == nil {
		return
	}

	if err := c.appendObject(ctx); err != nil {
		c.logger.Error("appendObject",
			"err", err,
			"bucket", c.bucket,
			"key", c.key,
		)
		return
	}

	if _, err := c.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(c.key),
		Body:   c.buf,
	}); err != nil {
		c.logger.Error("s3.PutObject",
			"err", err,
			"bucket", c.bucket,
			"key", c.key,
		)
		return
	}
}
