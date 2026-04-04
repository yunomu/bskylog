package handler

import (
	"context"
	"log/slog"

	"github.com/bluesky-social/indigo/api/bsky"
)

type Handler struct {
	logger *slog.Logger
}

func NewHandler(logger *slog.Logger) *Handler {
	return &Handler{
		logger: logger,
	}
}

type Request struct {
	DID   string                     `json:"did"`
	Posts []*bsky.FeedDefs_FeedViewPost `json:"posts"`
}

func (h *Handler) Handle(ctx context.Context, req *Request) error {
	h.logger.Info("IndexFunction received request", "did", req.DID, "num_posts", len(req.Posts))
	// ここに実際の処理を実装
	return nil
}
