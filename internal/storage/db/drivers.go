package db

import (
	"database/sql"

	_ "github.com/lib/pq"
)

type Queries interface {
	Save(key, value string) error
	Retrieve(key string) (value string, err error)
}

type Storage struct {
	*sql.DB
	Queries
}

func New(c *Config) *Storage {
	return &Storage{}
}

func newPostgresDB(dataSourceName string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, nil
}
