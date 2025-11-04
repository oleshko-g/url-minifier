// Package file is an implementation of [minifier.Storager]
package file

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"

	storageErrors "github.com/oleshko-g/url-minifier/internal/storage/errors"
)

// New returns a pointer to an opened [File] or an error
func New(c *Config) (file *File, err error) {
	fp, err := os.OpenFile(c.fpath.String(), os.O_RDWR|os.O_CREATE|os.O_APPEND, filePerm)
	if err != nil {
		slog.Error(fmt.Sprintf(" fp, err := os.OpenFile(c.fpath.String(), os.O_RDWR|os.O_CREATE|os.O_APPEND, filePerm) : %s", err))
		return nil, err
	}

	return &File{
		p:      fp,
		mux:    sync.RWMutex{},
		Config: c,
	}, nil
}

// File is a filesystem implementation of [minifier.Storager]
type File struct {
	mux sync.RWMutex
	p   *os.File
	*Config
}

// Close closes the underlying [os.File] of the [File]
func (f *File) Close() error {
	return f.p.Close()
}

type record struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Save saves the value under the key or return an error if key already exists
func (f *File) Save(key, value string) (err error) {
	f.mux.Lock()
	defer f.mux.Unlock()
	defer f.p.Sync()

	err = f.save(key, value)
	if err != nil {
		if errors.Is(err, storageErrors.ErrAlreadyExists) {
			slog.Warn(fmt.Sprintf("key %s already exists", key))
			return nil
		}
		return err
	}
	return
}

func (f *File) save(key, value string) (err error) {
	if _, err = f.retrieve(key); !errors.Is(err, storageErrors.ErrNotFound) {
		return storageErrors.ErrAlreadyExists
	}

	return json.NewEncoder(f.p).Encode(record{Key: key, Value: value})
}

// Retrieve returns a value stored in the [File] by a key or an [ErrNotFound]
func (f *File) Retrieve(key string) (value string, err error) {
	f.mux.RLock()
	defer f.mux.RUnlock()
	value, err = f.retrieve(key)
	if err != nil {
		slog.Error(fmt.Sprintf(" value, err = f.retrieve(key) %s", err.Error()))

		if errors.Is(err, storageErrors.ErrNotFound) {
			return "", fmt.Errorf("key %s doesn't exists", key)
		}
		return "", err
	}
	return value, nil
}

func (f *File) retrieve(key string) (value string, err error) {
	rr, err := f.newRecordReader()
	if err != nil {
		slog.Error(fmt.Sprintf(" rr, err := f.newRecordReader() %s", err.Error()))
		return "", err
	}
	defer rr.file.Close()
	var fr record
	for {
		err = rr.decoder.Decode(&fr)
		// safe to access [fr] because it's initialized to a zero value of [record]
		if fr.Key == key {
			return fr.Value, nil
		}
		if err != nil {
			slog.Error(fmt.Sprintf(" err = rr.decoder.Decode(&fr) %s", err.Error()))
			if errors.Is(err, io.EOF) {
				return "", storageErrors.ErrNotFound
			}
			return "", err
		}
		// set fr to zero value before the next Decode
		fr = record{}

	}
}

type recordReader struct {
	file    *os.File
	decoder *json.Decoder
}

func (f *File) newRecordReader() (*recordReader, error) {
	fp, err := os.OpenFile(f.p.Name(), os.O_RDONLY|os.O_CREATE, filePerm)
	if err != nil {
		slog.Error(fmt.Sprintf(" fp, err := os.OpenFile(f.p.Name(), os.O_RDONLY|os.O_CREATE, filePerm): %s", err.Error()))
		return nil, err
	}

	dep := json.NewDecoder(fp)

	return &recordReader{
		file:    fp,
		decoder: dep,
	}, nil
}

// defaults
const (
	DefaultFilepath path = "minifiedURLs.json"
)

// UNIX permissions
const (
	// Write Read _, Read _ _, Read _ _
	filePerm os.FileMode = 0o644
)
