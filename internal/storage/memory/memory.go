// Package memory is an in-memory implementation of [minifier.Storager]
package memory

import (
	"fmt"
	"sync"

	"github.com/oleshko-g/url-minifier/internal/storage"
	"github.com/oleshko-g/url-minifier/internal/storage/errors"
)

// NewStrRecords initializes and returns an in-memory implementation of [minifier.Storager]
func NewStrRecords() *strRecords { // revive:disable-line:unexported-return provides the interface to the caller
	records := &strRecords{
		mux: sync.RWMutex{},
		m:   make(map[string]string),
	}
	var i any = records
	if _, ok := i.(storage.Storager); !ok {
		return nil
	} else {
		return records
	}
}

// Ping is no-op for [strRecords]
func (s *strRecords) Ping() error {
	return nil
}

type strRecords struct {
	mux sync.RWMutex
	m   map[string]string
}

// FIXME: cannot use (*strRecords)(nil) (value of type *strRecords) as storage.Storager value in variable declaration: *strRecords does not implement storage.Storager (missing method SaveUserString) (compiler InvalidIfaceAssign)
// var _ storage.Storager = (*strRecords)(nil)

func (s *strRecords) Save(key, value string) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.m[key] = value
	return nil
}

// // // TODO: create userString struct, import, save in the map
// func (s *strRecords) SaveUserString(ctx context.Context, userID, key, value string) error {
// 	if ctx == nil {
// 		ctx = context.Background()
// 	}
// 	_, _, _, _ = ctx, userID, key, value
// 	return nil
// }

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
	v, ok := s.m[key]
	if !ok {
		return v, errors.ErrNotFound
	}
	return v, nil
}

func (s *strRecords) String() string {
	s.mux.RLock()
	defer s.mux.RUnlock()
	return fmt.Sprint(s.m)
}
