package file

// Config represents a [file.File] config
type Config struct {
	filePath path
}

// Path returns a pointer to an unexported [file.path] value
func (c *Config) Path() *path {
	return &c.filePath
}

// path respresents a valid file in a file system
type path string

// Set sets the [file.Config.Path()]
func (p *path) Set(s string) error {
	*p = path(s)
	return nil
}

// String return [file.Config.Path()] as a string value
func (p *path) String() string { return string(*p) }
