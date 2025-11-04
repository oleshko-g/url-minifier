package file

import (
	"fmt"
	"io/fs"
	"log/slog"
)

// Config represents a [file.File] config
type Config struct {
	fpath path
}

// Path returns a pointer to the [flag.Value] to set the [File]
func (c *Config) Path() *path { // revive:disable-line:unexported-return provides the interface to the caller
	return &c.fpath
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
