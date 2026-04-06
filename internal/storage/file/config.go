package file

import (
	"fmt"
	"io/fs"
	"log/slog"

	"github.com/oleshko-g/url-minifier/internal/config"
)

// NewConfig returns a new [Config] with the default values
func NewConfig() *Config {
	defaultFilepath := "minifiedURLs.json"

	return &Config{
		FilePath: config.Option[*path]{
			Name:        "f",
			Value:       new(path),
			Default:     defaultFilepath,
			Description: fmt.Sprintf("Default: `%s`. Set the file path for the file storage", defaultFilepath),
			Source:      config.SourceDefault,
		},
	}
}

// Config represents a [file.File] config
type Config struct {
	FilePath config.Option[*path]
}

// path represents a valid file in a file system
type path string

// Set sets the [file.Config.Path()]
func (p *path) Set(s string) error {
	if !fs.ValidPath(s) {
		slog.Warn(fmt.Sprintf("setting an invalid path: %s", s))
	}
	*p = path(s)
	return nil
}

// String return [file.Config.Path()] as a string value
func (p *path) String() string { return string(*p) }
