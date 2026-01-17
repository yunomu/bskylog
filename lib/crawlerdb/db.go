package crawlerdb

import (
	"context"
	"errors"
)

var ErrNotExists = errors.New("not exists")

type Crawler struct {
	Did       string
	Latest    string
	Timestamp string
}

type DB interface {
	Get(ctx context.Context, did string) (*Crawler, error)
	Put(ctx context.Context, crawler *Crawler) error
	Scan(ctx context.Context, f func(*Crawler) error) error
}
