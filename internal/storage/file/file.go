// Package file is an implementation of [minifier.Storager]
package file

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"sync"

	storageErrors "github.com/oleshko-g/url-minifier/internal/storage/errors"
)

// New returns a pointer to a [File] or—if a file [Config.Path().String()] is invalid—an error.
func New(c *Config) (file *File, err error) {
	if err = c.path.Set(c.path.String()); err != nil {
		log.Print(fmt.Errorf("c.filePath.Set(c.filePath.String() + fileName): %w", err))
		return nil, err
	}

	err = os.MkdirAll(c.path.String(), dirPerm)
	if err != nil {
		log.Print(fmt.Errorf("err = os.MkdirAll(c.filePath.String(), dirPerm): %w", err))
		return nil, err
	}

	fp, err := os.OpenFile(c.path.String()+fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, filePerm)
	if err != nil {
		log.Print(fmt.Errorf("fp, err := os.OpenFile(c.filePath.String()+fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, filePerm): %w", err))
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
	err = save(f, key, value)
	if err != nil {
		if errors.Is(err, storageErrors.ErrAlreadyExists) {
			return fmt.Errorf("value %s already exists", value)
		}
		return err
	}
	return
}

func save(f *File, key, value string) (err error) {
	f.mux.Lock()
	defer f.mux.Unlock()
	defer f.p.Sync()
	if _, err = f.retrieve(key); !errors.Is(err, storageErrors.ErrNotFound) {
		return storageErrors.ErrAlreadyExists
	}

	return json.NewEncoder(f.p).Encode(record{Key: key, Value: value})
}

// Retrieve returns a value stored in the [File] by a key or an [ErrNotFound]
func (f *File) Retrieve(key string) (value string, err error) {
	value, err = f.retrieve(key)
	if err != nil {
		if errors.Is(err, storageErrors.ErrNotFound) {
			return "", fmt.Errorf("key %s desn't exists", key)
		}
		return "", err
	}
	return value, nil
}

func (f *File) retrieve(key string) (value string, err error) {
	f.mux.RLock()
	defer f.mux.RUnlock()
	rr, err := f.newRecordReader()
	if err != nil {
		return "", err
	}
	defer rr.file.Close()
	var fr record
	for {
		err = rr.decoder.Decode(&fr)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return "", storageErrors.ErrNotFound
			}
			return "", err
		}

		if fr.Key == key {
			return fr.Value, nil
		}
	}
}

type recordReader struct {
	file    *os.File
	decoder *json.Decoder
}

func (f *File) newRecordReader() (*recordReader, error) {
	fp, err := os.OpenFile(f.Config.Path().String()+fileName, os.O_RDONLY|os.O_CREATE, filePerm)
	if err != nil {
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
	DefaultPath path   = "./files/"
	fileName    string = "minifiedURLs.json"
)

// UNIX persmissions
const (
	// Write Read Execute, Read _ Execute, Read _ Execute
	dirPerm os.FileMode = 0o755
	// Write Read _, Read _ _, Read _ _
	filePerm os.FileMode = 0o644
)
