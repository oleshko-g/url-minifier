// Package sql is the internal implementation of [database/sql]
package sql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	_ "github.com/lib/pq" // revive:disable-line:blank-imports registers the postgres driver
	"github.com/oleshko-g/url-minifier/internal/storage"
	"github.com/oleshko-g/url-minifier/internal/storage/db"
	query "github.com/oleshko-g/url-minifier/internal/storage/db/sql/queries"
	"github.com/oleshko-g/url-minifier/internal/storage/db/sql/schema"
	storageErrors "github.com/oleshko-g/url-minifier/internal/storage/errors"
)

// New configures and open a new connection to the db and returns a [Storage] or an error
func New(c *db.Config) (s *Storage, err error) {
	database, err := sql.Open(c.DSN().DriverName.String(), c.DSN().String())
	if err != nil {
		return nil, err
	}

	err = database.Ping()
	if err != nil {
		return nil, err
	}

	if err = schema.Up(c.DSN().DriverName, database); err != nil {
		return
	}

	return &Storage{
		db: database,
	}, nil
}

// Storage represents an internal implementation of [sql.DB]
type Storage struct {
	db *sql.DB
}

var _ storage.Storager = (*Storage)(nil)

// Ping exposes the Ping() method of the underlying [sql.DB]
func (s *Storage) Ping() error {
	return s.db.Ping()
}

// Save inserts value under key into the underlying db
//
// TODO: add tests
func (s *Storage) Save(key, value string) (err error) {
	err = s.save(key, value)
	if err != nil {
		if errors.Is(err, storageErrors.ErrAlreadyExists) {
			slog.Warn(fmt.Sprintf("key %s already exists", key))
		}
	}
	return err
}

// Save saves into the string_k_v db table
func (s *Storage) save(key, value string) error {
	row := s.db.QueryRow(
		query.InsertString,
		key, value, sql.Named("created_at", time.Now().UTC()), sql.NullTime{}, sql.NullTime{},
	)

	err := row.Scan(&key, &value)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return storageErrors.ErrAlreadyExists
		}
		return err
	}

	return nil
}

func (s *Storage) SaveUserString(ctx context.Context, userID, key, value string) error {
	if ctx == nil {
		ctx = context.Background()
	}

	err := s.saveUserString(ctx, userID, key, value)
	if err != nil {
		if errors.Is(err, storageErrors.ErrAlreadyExists) {
			slog.Warn(fmt.Sprintf("key %s already exists", key))
		}
	}
	return nil
}

func (s *Storage) saveUserString(ctx context.Context, userID, key, value string) error {
	row := s.db.QueryRow(
		query.InsertUserStrings,
		userID, key, value, sql.Named("created_at",
			time.Now().UTC()), sql.NullTime{}, sql.NullTime{},
	)

	err := row.Scan(&key, &value)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return storageErrors.ErrAlreadyExists
		}
		return err
	}

	return nil
}

// Retrieve selects value under key from the underlying db
func (s *Storage) Retrieve(key string) (value string, err error) {
	return s.retrieve(key)
}

func (s *Storage) retrieve(key string) (value string, err error) {
	row := s.db.QueryRow(query.SelectString, key)

	err = row.Scan(&value)
	if err != nil {
		_ = value
		if errors.Is(err, sql.ErrNoRows) {
			err = storageErrors.ErrNotFound
		}
		return "", err
	}

	return value, nil
}

// SaveList saves the slice of minified URLs coupled with their original URLs or returns an error
//
// TODO: add tests
func (s *Storage) SaveList(values []map[string]string) error {
	_ = values
	// TODL: write the implementation
	return nil
}
