// Package db is the base package for db servers
package db

import (
	"net/url"

	"github.com/oleshko-g/url-minifier/internal/config"
	storageErrors "github.com/oleshko-g/url-minifier/internal/storage/errors"
)

// NewConfig returns a new [Config] with default values
func NewConfig() Config {
	return Config{
		DSN: config.Option[*dataSource]{
			Name:        "d",
			Value:       new(dataSource),
			Description: "Set the sql db connection string",
			Default:     "",
			Source:      config.SourceDefault,
		},
	}
}

// Config represents a config of an SQL database
type Config struct {
	DSN config.Option[*dataSource]
}

// dataSource represent a valid Data Source
type dataSource struct {
	url string
	DriverName
	DefaultDSN string
	DBName     string
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
		DefaultDSN: defaultPostgesDSN(*url),
		DBName:     url.Path[1:], // slice off the leading '/'
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

func defaultPostgesDSN(url url.URL) string {
	url.Path = string(DriverNamePostgres)
	return url.String()
}
