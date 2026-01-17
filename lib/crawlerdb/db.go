package crawlerdb

import (
	"context"
	"errors"
)

var ErrNotExists = errors.New("not exists")

type Timestamp struct {
	Did       string
	LatestCid string
	Timestamp int
}

type DB interface {
	Get(ctx context.Context, did string) (*Timestamp, error)
	Put(ctx context.Context, ts *Timestamp) error
	Scan(ctx context.Context, f func(*Timestamp) error) error
}