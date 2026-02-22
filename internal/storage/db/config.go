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
	url string
	DriverName
	Host   string
	DBName string
}

// Set parses s and sets [DSN] and [Driver] or returns an error
func (d *dataSource) Set(s string) error {
	url, err := url.Parse(s)
	if err != nil {
		return err
	}

	if url.Scheme != string(DriverNamePostgres) {
		return storageErrors.ErrUnsupportedDataSource
	}

	*d = dataSource{
		url:        s,
		DriverName: DriverName(url.Scheme),
		Host:       url.Host,
		DBName:     url.Path,
	}

	return nil
}

func (d *dataSource) String() string {
	return d.url
}

// DriverName is a valid database driver name
type DriverName string

func (d DriverName) String() string {
	return string(d)
}

// Supported database drivers
const (
	DriverNamePostgres DriverName = "postgres"
)
