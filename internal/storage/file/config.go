package file

import (
	"fmt"
	"io/fs"
	"log/slog"
)

// Config represents a [file.File] config
type Config struct {
	path path
}

// Path returns a pointer to an unexported [file.path] value
func (c *Config) Path() *path {
	return &c.path
}

// path respresents a valid file in a file system
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
