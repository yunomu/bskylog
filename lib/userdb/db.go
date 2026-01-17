package userdb

import (
	"context"
	"errors"
)

type User struct {
	Did      string
	Handle   string
	Password string
	TimeZone int
}

var ErrNotExists = errors.New("not exists")

type DB interface {
	Get(ctx context.Context, did string) (*User, error)
	GetByHandle(ctx context.Context, handle string) (*User, error)
	Scan(ctx context.Context, f func(*User) error) error
	Put(ctx context.Context, user *User) error
}
