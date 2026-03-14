// Package file is an implementation of [minifier.Storager]
package file

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/oleshko-g/url-minifier/internal/storage"
	storageErrors "github.com/oleshko-g/url-minifier/internal/storage/errors"
)

// New returns a pointer to an opened [File] or an error
func New(c *Config) (file storage.StoragePinger, err error) {
	fp, err := os.OpenFile(c.fpath.String(), os.O_RDWR|os.O_CREATE|os.O_APPEND, filePerm)
	if err != nil {
		return nil, err
	}

	return storage.NewStoragePingerNoOp(
			&File{
				p:      fp,
				mux:    sync.RWMutex{},
				Config: c,
			}),
		nil
}

// File is a filesystem implementation of [minifier.Storager]
// TODO: refactor into "FileStorage" and separate the [File]
type File struct {
	mux sync.RWMutex
	p   *os.File
	*Config
}

var _ storage.Storager = (*File)(nil)

// Close closes the underlying [os.File] of the [File]
func (f *File) Close() error {
	return f.p.Close()
}

type record struct {
	UserID     string     `json:"user_id"`
	Key        string     `json:"key"`
	Value      string     `json:"value"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  *time.Time `json:"updated_at"`
	DeletedtAt *time.Time `json:"deleted_at"`
}

func (r record) isDeleted() bool {
	if r.DeletedtAt == nil {
		return false
	}
	return time.Now().UTC().After(*r.DeletedtAt)
}

// Save saves the value under the key or return an error if key already exists
//
// TODO: add tests
func (f *File) Save(key, value string) error {
	f.mux.Lock()
	defer f.mux.Unlock()
	// TODO: sync every 100 ms instead of every write
	// TODO: add in-memory storage and update on every Sync()
	defer f.p.Sync()

	err := f.save(key, value)
	if err != nil {
		if errors.Is(err, storageErrors.ErrAlreadyExists) {
			slog.Warn(fmt.Sprintf("key %s already exists", key))
			return nil
		}
		return err
	}
	return nil
}

func (f *File) save(key, value string) error {
	if _, err := f.retrieve(key); !errors.Is(err, storageErrors.ErrNotFound) {
		return storageErrors.ErrAlreadyExists
	}

	return json.NewEncoder(f.p).Encode(record{Key: key, Value: value, CreatedAt: time.Now().UTC()})
}

// SaveUserString appends a [storage.UserString] to the file storage
func (f *File) SaveUserString(ctx context.Context, us storage.UserString) error {
	_ = ctx

	f.mux.Lock()
	defer f.mux.Unlock()
	err := f.saveUserString(us)
	if err != nil {
		if errors.Is(err, storageErrors.ErrAlreadyExists) {
			slog.Warn(fmt.Sprintf("key %s already exists", us.Key))
			return nil
		}
		return err
	}
	return nil
}

func (f *File) saveUserString(us storage.UserString) error {

	if _, err := f.retrieve(us.Key); !errors.Is(err, storageErrors.ErrNotFound) {
		return storageErrors.ErrAlreadyExists
	}
	return json.NewEncoder(f.p).Encode(record{UserID: us.UserID, Key: us.Key, Value: us.Value})
}

// SaveList saves the slice of minified URLs coupled with their original URLs or returns an error
//
// TODO: add tests
func (f *File) SaveList(values []map[string]string) error {
	_ = values
	// TODL: write the implementation
	return nil
}

// Retrieve returns a value stored in the [File] by a key or an [ErrNotFound]
//
// TODO: add tests
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
	// TODO: lookup in-memory storage first
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

// RetrieveUserStrings takes userID and returns a slice of [storage.UserString]'s or an error
func (f *File) RetrieveUserStrings(ctx context.Context, userID string) ([]storage.UserString, error) {
	_ = ctx
	f.mux.RLock()
	defer f.mux.RUnlock()

	rr, err := f.newRecordReader()
	if err != nil {
		return nil, err
	}
	var urs []storage.UserString
	for {
		var fr record
		err = rr.decoder.Decode(&fr)
		if fr.UserID == userID {
			urs = append(urs, storage.UserString{UserID: userID, Key: fr.Key, Value: fr.Value})
		}

		if err != nil {
			if errors.Is(err, io.EOF) {
				if len(urs) == 0 {
					return nil, storageErrors.ErrNotFound
				}

				break
			}

			return nil, err
		}
	}

	return urs, nil
}

func (f *File) retrieveUserStrings(userID string) (key, value string, err error) { //revive:disable-line:unused-parameter userID is used for comparison only

	return "", "", nil
}

// MarkDeletedUserString is the file implementation
//
// TODO: retrieve a value, set deletedAt if it's not
func (f *File) MarkDeletedUserString(ctx context.Context, userID string, key string) error {
	_, _, _ = ctx, userID, key
	return nil
}

// RetrieveUserString is the file implementation
func (f *File) RetrieveUserString(ctx context.Context, key string) (storage.UserString, error) {
	_ = ctx
	f.mux.RLock()
	defer f.mux.RUnlock()

	rr, err := f.newRecordReader()
	if err != nil {
		return storage.UserString{}, err
	}
	for {
		var fr record
		err = rr.decoder.Decode(&fr)
		if fr.Key == key {
			return storage.UserString{UserID: fr.UserID,
				Key:     fr.Key,
				Value:   fr.Value,
				Deleted: fr.isDeleted(),
			}, nil
		}

		if err != nil {
			if errors.Is(err, io.EOF) {

				return storage.UserString{}, storageErrors.ErrNotFound
			}
			return storage.UserString{}, err
		}
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
