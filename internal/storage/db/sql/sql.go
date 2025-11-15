// Package sql is the internal implementation of [database/sql]
package sql

import (
	"database/sql"
	"time"

	_ "github.com/lib/pq" // revive:disable-line:blank-imports registers the postgres driver
	"github.com/oleshko-g/url-minifier/internal/storage/db"
	query "github.com/oleshko-g/url-minifier/internal/storage/db/sql/queries"
)

// Storage represents an internal implementation of [sql.DB]
type Storage struct {
	db *sql.DB
}

// Ping exposes the Ping() method of the underlying [sql.DB]
func (s *Storage) Ping() error {
	return s.db.Ping()
}

// Save inserts value under key into the underlying db
func (s *Storage) Save(key, value string) error {
	return s.save(key, value)
}

// Save saves into the string_k_v db table
func (s *Storage) save(key, value string) error {
	_, err := s.db.Exec(
		query.InsertString,
		key, value, sql.Named("created_at", time.Now().UTC()), sql.NullTime{}, sql.NullTime{},
	)
	if err != nil {
		return err
	}

	return nil
}

// Retrieve selects value under key from the underlying db
func (s *Storage) Retrieve(key string) (value string, err error) {
	// TODO: write func selectKeyValue(key string) (value string, err error)
	_, _, _ = key, value, err
	return "", nil
}

func (s *Storage) retrieve(key string) (value string, err error) {
	return "", nil
}

// New configures and open a new connection to the db and returns a [Storage] or an error
func New(c *db.Config) (s *Storage, err error) {
	db, err := sql.Open(c.DSN().Driver(), c.DSN().String())
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return &Storage{
		db: db,
	}, nil
}
