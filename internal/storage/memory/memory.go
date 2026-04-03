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
func NewStrRecords() storage.PingerCloser { // revive:disable-line:unexported-return provides the interface to the caller
	uks := make(userKeys)
	records := storage.NewNoOpPingerClose(
		&strRecords{
			mux:      sync.RWMutex{},
			userKeys: uks,
			values:   make(values),
		})

	return records
}

type (
	strRecords struct {
		mux sync.RWMutex
		userKeys
		values
	}

	user string
	key  = string

	userKeys = map[user][]key
	values   = map[key]strValue

	strValue struct {
		userID    user
		str       string
		createdAt time.Time
		updatedAt *time.Time
		deletedAt *time.Time
	}
)

func (v strValue) isDeleted() bool {
	if v.deletedAt == nil {
		return false
	}
	return time.Now().UTC().After(*v.deletedAt)
}

var _ storage.Storager = (*strRecords)(nil)

func (s *strRecords) Save(key, value string) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	err := s.save(key, value)
	if err != nil {
		return err
	}
	return nil
}

func (s *strRecords) save(key, v string) error {
	_, ok := s.values[key]
	if ok {
		return storageErrors.ErrAlreadyExists
	}

	s.values[key] = strValue{str: v, createdAt: time.Now().UTC(), updatedAt: nil, deletedAt: nil}
	return nil
}

// // TODO: create userString struct, import, save in the map
func (s *strRecords) SaveUserString(ctx context.Context, us storage.UserString) error {
	_ = ctx

	s.mux.Lock()
	defer s.mux.Unlock()

	err := s.saveUserString(user(us.UserID), us.Key, us.Value)
	if errors.Is(err, storageErrors.ErrAlreadyExists) {
		slog.Warn(fmt.Sprintf("key %s already exists", us.Key))
	}

	return nil
}

func (s *strRecords) saveUserString(u user, k key, str string) error {
	if _, ok := s.values[k]; ok {
		return storageErrors.ErrAlreadyExists
	}

	s.values[k] = strValue{
		userID:    u,
		str:       str,
		createdAt: time.Now().UTC(),
	}

	uks := s.userKeys[u]
	s.userKeys[u] = append(uks, k)
	return nil
}

// SaveList saves the slice of minified URLs coupled with their original URLs or returns an error
//
// TODO: add tests
func (s *strRecords) SaveList(values []map[string]string) error {
	_ = values
	// TODO: write the implementation
	return nil
}

func (s *strRecords) Retrieve(key string) (value string, err error) {
	s.mux.RLock()
	defer s.mux.RUnlock()
	v, ok := s.values[key]
	if !ok {
		return "", storageErrors.ErrNotFound
	}
	return v.str, nil
}

func (s *strRecords) RetrieveUserStrings(ctx context.Context, userID string) ([]storage.UserString, error) {
	_ = ctx

	s.mux.RLock()
	defer s.mux.RUnlock()

	uks, ok := s.userKeys[user(userID)]
	if !ok {
		return []storage.UserString{}, nil
	}

	var uss []storage.UserString
	for _, uk := range uks {
		v, err := s.retrieveUserString(uk)
		if err != nil {
			return nil, err
		}
		uss = append(uss, v)
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

func (s *strRecords) RetrieveUserString(ctx context.Context, key string) (storage.UserString, error) {
	_ = ctx
	s.mux.RLock()
	defer s.mux.RUnlock()

	return s.retrieveUserString(key)
}

func (s *strRecords) retrieveUserString(k key) (storage.UserString, error) {
	v, ok := s.values[k]
	if !ok {
		return storage.UserString{}, storageErrors.ErrNotFound
	}

	return storage.UserString{
		UserID:  string(v.userID),
		Key:     k,
		Value:   v.str,
		Deleted: v.isDeleted(),
	}, nil
}
