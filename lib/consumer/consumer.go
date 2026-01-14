package consumer

import (
	"context"

	"github.com/bluesky-social/indigo/api/bsky"
)

type Consumer interface {
	Consume(ctx context.Context, post *bsky.FeedDefs_FeedViewPost) error
	Close()
}
