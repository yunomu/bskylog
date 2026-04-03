package storage

import (
	"context"

	"github.com/bluesky-social/indigo/api/bsky"
)

// Scanner is an interface for scanning posts.
type Scanner interface {
	Scan(ctx context.Context, f func(key string, position int, post *bsky.FeedDefs_FeedViewPost) error) error
}
