package consumer

import (
	"context"
	"encoding/json"
	"io"

	"github.com/bluesky-social/indigo/api/bsky"
)

type JSONRecord struct {
	encoder *json.Encoder
}

var _ Consumer = (*JSONRecord)(nil)

func NewJSONRecord(w io.Writer) *JSONRecord {
	return &JSONRecord{
		encoder: json.NewEncoder(w),
	}
}

func (c *JSONRecord) Consume(ctx context.Context, post *bsky.FeedDefs_FeedViewPost) error {
	return c.encoder.Encode(post)
}

func (c *JSONRecord) Close(_ context.Context) error {
	return nil
}
