package consumer

import (
	"bytes"
	"context"
	"io"
	"sync/atomic"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	smithy "github.com/aws/smithy-go"
	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/indigo/lex/util"
	"github.com/stretchr/testify/assert"
)

// MockS3Client is a mock implementation of the S3Client interface.
type MockS3Client struct {
	PutObjectFunc func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	GetObjectFunc func(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
}

func (m *MockS3Client) PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	if m.PutObjectFunc != nil {
		return m.PutObjectFunc(ctx, params, optFns...)
	}
	return &s3.PutObjectOutput{}, nil
}

func (m *MockS3Client) GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	if m.GetObjectFunc != nil {
		return m.GetObjectFunc(ctx, params, optFns...)
	}
	return nil, &smithy.OperationError{
		Err: &types.NoSuchKey{},
	}
}

func TestDailyJSONRecordS3FirstCallOnce(t *testing.T) {
	ctx := context.Background()

	var firstCallCount atomic.Int32

	mockS3 := &MockS3Client{
		PutObjectFunc: func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
			return &s3.PutObjectOutput{}, nil
		},
		GetObjectFunc: func(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
			if *params.Key == "base/2023/01/index" || *params.Key == "base/2023/02/index" {
				return nil, &smithy.OperationError{
					Err: &types.NoSuchKey{},
				}
			}
			return &s3.GetObjectOutput{
				Body: io.NopCloser(bytes.NewReader([]byte{})),
			}, nil
		},
	}

	c := NewDailyJSONRecordS3(
		mockS3,
		"test-bucket",
		"base",
		time.UTC,
		SetDailyJSONRecordS3FirstValueFunc(func(ts int64, cid string) {
			firstCallCount.Add(1)
		}),
	)

	post1 := &bsky.FeedDefs_FeedViewPost{
		Post: &bsky.FeedDefs_PostView{
			Cid: "cid1",
			Record: &util.LexiconTypeDecoder{
				Val: &bsky.FeedPost{
					CreatedAt: time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC).Format(time.RFC3339Nano),
					Text:      "test1",
				},
			},
		},
	}
	post2 := &bsky.FeedDefs_FeedViewPost{
		Post: &bsky.FeedDefs_PostView{
			Cid: "cid2",
			Record: &util.LexiconTypeDecoder{
				Val: &bsky.FeedPost{
					CreatedAt: time.Date(2023, 1, 1, 11, 0, 0, 0, time.UTC).Format(time.RFC3339Nano),
					Text:      "test2",
				},
			},
		},
	}
	post3 := &bsky.FeedDefs_FeedViewPost{
		Post: &bsky.FeedDefs_PostView{
			Cid: "cid3",
			Record: &util.LexiconTypeDecoder{
				Val: &bsky.FeedPost{
					CreatedAt: time.Date(2023, 2, 1, 12, 0, 0, 0, time.UTC).Format(time.RFC3339Nano), // Change date to trigger ensureStream.
					Text:      "test3",
				},
			},
		},
	}

	err := c.Consume(ctx, post1)
	assert.NoError(t, err)

	err = c.Consume(ctx, post2)
	assert.NoError(t, err)

	err = c.Consume(ctx, post3) // ensureStream is called, starting a new stream.
	assert.NoError(t, err)

	err = c.Close(ctx)
	assert.NoError(t, err)

	assert.Equal(t, int32(1), firstCallCount.Load(), "first should be called exactly once")
}
