package index

import (
	"context"
	"errors"
	"log/slog"

	"gorm.io/gorm"

	"github.com/bluesky-social/indigo/api/bsky"
)

type Gorm struct {
	db     *gorm.DB
	logger *slog.Logger
}

type GormOption func(*Gorm)

func GormOptionLogger(logger *slog.Logger) GormOption {
	return func(g *Gorm) {
		if logger == nil {
			g.logger = slog.Default()
		} else {
			g.logger = logger
		}
	}
}

func NewGorm(db *gorm.DB, opts ...GormOption) *Gorm {
	g := &Gorm{
		db:     db,
		logger: slog.Default(),
	}

	for _, opt := range opts {
		opt(g)
	}

	err := db.AutoMigrate(&Record{})
	if err != nil {
		g.logger.Error("failed to auto migrate Record table", "err", err)
		panic(err)
	}
	return g
}

func (s *Gorm) Put(ctx context.Context, key string, position int, post *bsky.FeedDefs_FeedViewPost) error {
	rec := ToRecord(key, position, post)
	if rec == nil {
		err := errors.New("unexpected post")
		s.logger.Error("failed to create record", "key", key, "position", position, "err", err)
		return err
	}

	if err := s.db.WithContext(ctx).Create(rec).Error; err != nil {
		s.logger.Error("failed to put record into database", "key", key, "cid", rec.Cid, "err", err)
		return err
	}
	return nil
}
