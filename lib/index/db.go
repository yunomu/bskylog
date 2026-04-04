package index

import (
	"context"
	"errors"

	"github.com/bluesky-social/indigo/api/bsky"
	"gorm.io/gorm"
)

type Gorm struct {
	db *gorm.DB
}

func NewGorm(db *gorm.DB) *Gorm {
	return &Gorm{db: db}
}

func (s *Gorm) Put(ctx context.Context, key string, position int, post *bsky.FeedDefs_FeedViewPost) error {
	rec := ToRecord(key, position, post)
	if rec == nil {
		return errors.New("unexpected post")
	}

	return s.db.WithContext(ctx).Create(rec).Error
}
