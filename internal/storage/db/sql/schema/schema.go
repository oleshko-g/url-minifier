// Package schema is the package to run SQL migrations
package schema

import (
	"database/sql"
	"embed"
	"errors"
	"time"

	"github.com/oleshko-g/url-minifier/internal/storage/db"
	"github.com/pressly/goose/v3"
)

//go:embed psql/*.sql
var psqlMigrations embed.FS

// Up runs
func Up(d db.DriverName, database *sql.DB) error {
	if err := goose.SetDialect(d.String()); err != nil {
		return err
	}

	var dir string
	if d == db.DriverNamePostgres {
		goose.SetBaseFS(psqlMigrations)
		dir = "psql"
	} else {
		return errors.New("driver is not supported")
	}

	if err := goose.Up(database, dir); err != nil {
		return err
	}
	return nil
}

// UserString is the struct to scan data from SQL queries to strings table
type UserString struct {
	UserID    string
	Value     string
	DeletedAt *time.Time
}

// IsDeleted reports if a value's deleted_at field is in the past so it's marked as deleted
func (us UserString) IsDeleted() bool {
	if us.DeletedAt != nil {
		if time.Now().After(*us.DeletedAt) {
			return true
		}
	}
	return false
}
