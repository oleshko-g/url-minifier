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
	SaveUserString(ctx context.Context, user_id, key, value string) error
}

type userString struct {
	userID, key, value string
}
