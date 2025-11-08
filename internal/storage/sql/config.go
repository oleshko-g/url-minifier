// Package sql is internal implementation of SQL
package sql

// Config represents a config of an SQL database
type Config struct {
	dbConn
}

// DBConn returns a pointer to the [flag.Value] to set the database connection string
func (c *Config) DBConn() *dbConn { // revive:disable-line:unexported-return provides the interface to the caller
	return &c.dbConn
}

type dbConn string

func (d *dbConn) Set(s string) error {
	*d = dbConn(s)
	return nil
}

func (d *dbConn) String() string {
	return string(*d)
}
