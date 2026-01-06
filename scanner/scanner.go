package scanner

import (
	"context"
	"log/slog"

	"github.com/bluesky-social/indigo/api/bsky"
	lexutil "github.com/bluesky-social/indigo/lex/util"
)

type Scanner interface {
	Scan(ctx context.Context, filter string, includePins bool, f func([]*bsky.FeedDefs_FeedViewPost) error) error
}

type XRPCScanner struct {
	client lexutil.LexClient
	actor  string
	limits int64

	logger *slog.Logger
}

var _ Scanner = (*XRPCScanner)(nil)

type XRPCScannerOption func(*XRPCScanner)

func SetLogger(l *slog.Logger) XRPCScannerOption {
	return func(s *XRPCScanner) {
		if l == nil {
			s.logger = slog.Default()
		} else {
			s.logger = l
		}
	}
}

func NewXRPCScanner(client lexutil.LexClient, actor string, opts ...XRPCScannerOption) *XRPCScanner {
	ret := &XRPCScanner{
		client: client,
		actor:  actor,
		limits: 100,
	}
	for _, f := range opts {
		f(ret)
	}
	return ret
}

func (s *XRPCScanner) Scan(ctx context.Context, filter string, includePins bool, f func([]*bsky.FeedDefs_FeedViewPost) error) error {
	var cursor string
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// do nothing
		}

		feed, err := bsky.FeedGetAuthorFeed(ctx, s.client, s.actor, cursor, filter, includePins, s.limits)
		if err != nil {
			s.logger.Error("FeedGetAuthorFeed",
				"actor", s.actor,
				"cursor", cursor,
				"filter", filter,
				"includePins", includePins,
				"limits", s.limits,
			)
			return err
		}

		if err := f(feed.Feed); err != nil {
			return err
		}

		if feed.Cursor != nil && *feed.Cursor != "" {
			cursor = *feed.Cursor
		} else {
			return nil
		}
	}
}
