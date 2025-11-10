// Package db is the base package for db servers
package db

import (
	"net/url"

	storageErrors "github.com/oleshko-g/url-minifier/internal/storage/errors"
)

// Config represents a config of an SQL database
type Config struct {
	dataSource
}

// DSN returns a pointer to the [flag.Value] to set the database source name
func (c *Config) DSN() *dataSource { // revive:disable-line:unexported-return provides the interface to the caller
	return &c.dataSource
}

// dataSource represent a valid Data Source
type dataSource struct {
	name string
	driver
}

func (d *dataSource) Set(s string) error {
	url, err := url.Parse(s)
	if err != nil {
		_ = url
		return err
	}

	if url.Scheme != string(postgres) {
		_ = url
		return storageErrors.ErrUnsupportedDataSource
	}
	d.driver = driver(url.Scheme)

	d.name = url.String()

	return nil
}

func (d *dataSource) String() string {
	return d.name
}

func (d *dataSource) Driver() string {
	return string(d.driver)
}

type driver string

const postgres driver = "postgres"
