// Package memory is an in-memory implementation of [minifier.Storager]
package memory

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/oleshko-g/url-minifier/internal/storage"
	storageErrors "github.com/oleshko-g/url-minifier/internal/storage/errors"
)

// NewStrRecords initializes and returns an in-memory implementation of [minifier.Storager]
func NewStrRecords() *strRecords { // revive:disable-line:unexported-return provides the interface to the caller
	uks := make(userKeys)
	records := &strRecords{
		mux:      sync.RWMutex{},
		userKeys: uks,
		values:   make(values),
	}

	return records
}

// Ping is no-op for [strRecords]
func (s *strRecords) Ping() error {
	return nil
}

type (
	strRecords struct {
		mux sync.RWMutex
		userKeys
		values
	}

	user string
	key  = string

	userKeys = map[user][]userKey
	values   = map[key]string

	userKey struct {
		key
		createdAt time.Time
		updatedAt time.Time
		deletedAt *time.Time
	}
)

var _ storage.Storager = (*strRecords)(nil)

func (s *strRecords) Save(key, value string) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.save(key, value)
	return nil
}

func (s *strRecords) save(key, value string) error {
	_, ok := s.values[key]
	if ok {
		return storageErrors.ErrAlreadyExists
	}

	s.values[key] = value
	return nil
}

// // TODO: create userString struct, import, save in the map
func (s *strRecords) SaveUserString(ctx context.Context, us storage.UserString) error {
	_ = ctx

	s.mux.Lock()
	defer s.mux.Unlock()

	if err := s.save(us.Key, us.Value); err != nil {
		if errors.Is(err, storageErrors.ErrAlreadyExists) {
			slog.Warn(fmt.Sprintf("key %s already exists", us.Key))
		}
		return err
	}

	urs := s.userKeys[user(us.UserID)]
	urs = append(urs, userKey{
		key:       us.Key,
		createdAt: time.Now().UTC(),
		updatedAt: time.Now().UTC(),
		deletedAt: nil,
	})

	s.userKeys[user(us.UserID)] = urs

	return nil
}

// SaveList saves the slice of minified URLs coupled with their original URLs or returns an error
//
// TODO: add tests
func (s *strRecords) SaveList(values []map[string]string) error {
	_ = values
	// TODL: write the implementation
	return nil
}

func (s *strRecords) Retrieve(key string) (value string, err error) {
	s.mux.RLock()
	defer s.mux.RUnlock()
	v, ok := s.values[key]
	if !ok {
		return "", storageErrors.ErrNotFound
	}
	return v, nil
}

func (s *strRecords) RetrieveUserStrings(ctx context.Context, userID string) ([]storage.UserString, error) {
	_ = ctx

	s.mux.RLock()
	defer s.mux.RUnlock()

	urs, ok := s.userKeys[user(userID)]
	if !ok {
		return []storage.UserString{}, nil
	}

	var uss []storage.UserString
	for _, ur := range urs {
		us := storage.UserString{
			UserID: userID,
			Key:    ur.key,
			Value:  s.values[ur.key],
		}
		uss = append(uss, us)
	}
	return uss, nil
}

func (s *strRecords) String() string {
	s.mux.RLock()
	defer s.mux.RUnlock()
	return fmt.Sprint(s.userKeys)
}

func (s *strRecords) MarkDeletedUserString(ctx context.Context, userID string, key string) error {
	_, _, _ = ctx, userID, key
	return nil
}
