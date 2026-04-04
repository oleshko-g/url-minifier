// Package storage is the package intended to be imported by minifier package and every storage implementation
package storage

import (
	"context"
	"io"
)

//go:generate moq -pkg mockStorage -out ../mock/storage/storage.go . PingerCloserCounter
// PingerCloserCounter is the expected implementation of a [PingerCloser] with a [Counter] for the URL minifier
type PingerCloserCounter interface {
	PingerCloser
	Counter
}

type Counter interface {
	CountUserStrings(context.Context) (int, error)
	CountUsers(context.Context) (int, error)
}

// Storager is the expected implementation of storage for the URL minifier
type Storager interface {
	Save(key, value string) error
	SaveList(values []map[string]string) error
	Retrieve(key string) (value string, err error)
	SaveUserString(ctx context.Context, us UserString) error
	RetrieveUserStrings(ctx context.Context, userID string) ([]UserString, error)
	RetrieveUserString(ctx context.Context, key string) (UserString, error)
	MarkDeletedUserString(ctx context.Context, userID string, key string) error
}

// Pinger is the expected implementation of a pinger for the URL minifier
type Pinger interface {
	Ping() error
}

// PingerCloser is the expected implementation of storage with a [Pinger] for the URL minifier
type PingerCloser interface {
	Storager
	Pinger
	io.Closer
}

// PingerNoOp is a wraper struct to implement a noop [Pinger]
type PingerNoOp struct {
	Storager
}

// NewNoOpPingerClose wraps a [Storager] and return a [PingerCloser]
func NewNoOpPingerClose(storager Storager) PingerCloser {
	return &PingerNoOp{Storager: storager}
}

// Ping is no-op for [Ping]
func (p *PingerNoOp) Ping() error {
	return nil
}

func (p *PingerNoOp) Close() error {
	return nil
}

// UserString is the structure used in [storage.Storager] implementations
type UserString struct {
	UserID, Key, Value string
	Deleted            bool
}
