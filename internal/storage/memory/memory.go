package memory

import (
	"fmt"
	"sync"

	"github.com/oleshko-g/url-minifier/internal/storage/errors"
)

func NewStrRecords() *strRecords {
	return &strRecords{
		mux: sync.RWMutex{},
		m:   make(map[string]string),
	}
}

type strRecords struct {
	mux sync.RWMutex
	m   map[string]string
}

func (s *strRecords) Save(key, value string) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.m[key] = value
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
