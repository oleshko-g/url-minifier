package db

import (
	"database/sql"

	_ "github.com/lib/pq" // revive:disable-line:blank-imports registers the postgres driver
)

// Queries defines the possible queries to a database
type Queries interface {
	Save(key, value string) error
	Retrieve(key string) (value string, err error)
}

// Storage represents an internal implementation of [sql.DB]
type Storage struct {
	db *sql.DB
	Queries
}

// Ping exposes the Ping() method of the underlying [sql.DB]
func (s *Storage) Ping() error {
	return s.db.Ping()
}

// New configures and open a new connection to the db and returns a [Storage] or an error
func New(c *Config) (s *Storage, err error) {
	db, err := sql.Open(string(c.dataSource.driver), c.DSN().String())
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return &Storage{
		db:      db,
		Queries: nil,
	}, nil
}
