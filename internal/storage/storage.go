// Package storage is the package intended to be imported by minifier package and every storage implementation
package storage

import "context"

// Storager is the expected implementation of storage for the URL minifier
//
//go:generate moq -pkg mockStorage -out ../mock/storage/storage.go . Storager
type Storager interface {
	Save(key, value string) error
	SaveList(values []map[string]string) error
	Retrieve(key string) (value string, err error)
	Ping() error
	SaveUserString(ctx context.Context, us UserString) error
	RetrieveUserStrings(ctx context.Context, userID string) ([]UserString, error)
	MarkDeletedUserString(ctx context.Context, userID string, key string) error
}

// UserString is the structure used in [storage.Storager] implementations
type UserString struct {
	UserID, Key, Value string
	Deleted            bool
}
