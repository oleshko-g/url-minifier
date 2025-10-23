package file

// Config represents a [file.File] config
type Config struct {
	p path
}

// Path returns a pinter to an unexported [file.path] value
func (c *Config) Path() *path {
	return &c.p
}

type path string

// Set sets the [file.Config.Path()]
func (p *path) Set(s string) error {
	*p = path(s)
	return nil
}

// String return [file.Config.Path()] as a string value
func (p *path) String() string { return string(*p) }
