// Package sql is the internal implementation of [database/sql]
package sql // revive:disable-line:var-naming see the package description.

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	_ "github.com/lib/pq" // revive:disable-line:blank-imports registers the postgres driver
	"github.com/oleshko-g/url-minifier/internal/storage"
	"github.com/oleshko-g/url-minifier/internal/storage/db"
	query "github.com/oleshko-g/url-minifier/internal/storage/db/sql/queries"
	"github.com/oleshko-g/url-minifier/internal/storage/db/sql/schema"
	storageErrors "github.com/oleshko-g/url-minifier/internal/storage/errors"
)


// New configures and open a new connection to the db and returns a [Storage] or an error
func New(dbCfg db.Config) (*Storage, error) {
	db, err := connectDB(string(dbCfg.DriverName), dbCfg.DSN().String())
	if err != nil {
		err = createDB(dbCfg)
		if err != nil {
			return nil, err
		}

		db, err = connectDB(string(dbCfg.DriverName), dbCfg.DSN().String())
		if err != nil {
			return nil, err
		}

	}

	if err = schema.Up(dbCfg.DSN().DriverName, db); err != nil {
		return nil, err
	}

	return &Storage{Config: dbCfg, db: db}, nil
}

// Storage represents an internal implementation of [sql.DB]
type Storage struct {
	db *sql.DB
	db.Config
}

var _ storage.Storager = (*Storage)(nil)

// Ping exposes the Ping() method of the underlying [sql.DB]
func (s *Storage) Ping() error {
	return s.db.Ping()
}

// Save inserts value under key into the underlying db
//
// TODO: add tests
func (s *Storage) Save(key, value string) (err error) {
	err = s.save(key, value)
	if err != nil {
		if errors.Is(err, storageErrors.ErrAlreadyExists) {
			slog.Warn(fmt.Sprintf("key %s already exists", key))
		}
	}
	return err
}

// Save saves into the string_k_v db table
func (s *Storage) save(key, value string) error {
	row := s.db.QueryRow(
		query.InsertString,
		key, value, sql.Named("created_at", time.Now().UTC()), sql.NullTime{}, sql.NullTime{},
	)

	err := row.Scan(&key, &value)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return storageErrors.ErrAlreadyExists
		}
		return err
	}

	return nil
}

// SaveUserString saves a [storage.UserString] in the database
func (s *Storage) SaveUserString(ctx context.Context, us storage.UserString) error {
	if ctx == nil {
		ctx = context.Background()
	}

	err := s.saveUserString(ctx, us)
	if err != nil {
		if errors.Is(err, storageErrors.ErrAlreadyExists) {
			slog.Warn(fmt.Sprintf("key %s already exists", us.Key))
		}
		return err
	}
	return nil
}

func (s *Storage) saveUserString(ctx context.Context, us storage.UserString) error {
	row := s.db.QueryRowContext(ctx,
		query.InsertUserString,
		us.UserID,
		us.Key,
		us.Value,
		sql.Named("created_at", time.Now().UTC()),
		sql.NullTime{},
		sql.NullTime{},
	)

	err := row.Scan(&us.Key, &us.Value)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return storageErrors.ErrAlreadyExists
		}
		return err
	}

	return nil
}

// Retrieve selects value under key from the underlying db
func (s *Storage) Retrieve(key string) (value string, err error) {
	return s.retrieve(key)
}

func (s *Storage) retrieve(key string) (value string, err error) {
	row := s.db.QueryRow(query.SelectString, key)

	err = row.Scan(&value)
	if err != nil {
		_ = value
		if errors.Is(err, sql.ErrNoRows) {
			err = storageErrors.ErrNotFound
		}
		return "", err
	}

	return value, nil
}

// RetrieveUserStrings takes userID and return a slice of [storage.UserString]'s
//
// TODO: add test
func (s *Storage) RetrieveUserStrings(ctx context.Context, userID string) ([]storage.UserString, error) {
	var err error

	if userID == "" {
		err = fmt.Errorf("%w: %s", storageErrors.ErrEmptyParameter, "userID")
		return nil, err
	}

	rows, err := s.db.QueryContext(ctx, query.SelectUserStrings, userID)
	if err != nil {
		return nil, err
	}

	var uss []storage.UserString
	for rows.Next() {
		var us storage.UserString
		err := rows.Scan(&us.UserID, &us.Key, &us.Value)
		if err != nil {
			return nil, err
		}

		uss = append(uss, us)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return uss, nil
}

// SaveList saves the slice of minified URLs coupled with their original URLs or returns an error
//
// TODO: add tests
func (s *Storage) SaveList(values []map[string]string) error {
	_ = values
	// TODL: write the implementation
	return nil
}

// MarkDeletedUserString sets deleted_at. If the user isn't the ownder it returns [storageErrors.AccessDenifed]
func (s *Storage) MarkDeletedUserString(ctx context.Context, userID string, key string) error {
	var err error

	dus, err := s.retrieveUserString(ctx, key)
	if errors.Is(err, sql.ErrNoRows) {
		err = fmt.Errorf("%w: by key \"%s\"", storageErrors.ErrNotFound, key)
		return err
	}

	if userID != dus.UserID {
		err = fmt.Errorf("%w: userID \"%s\" %s", storageErrors.ErrAccessDenied, userID, "isn't the owner of data")
		return err
	}

	if dus.IsDeleted() {
		return nil
	}

	err = s.updateStringDeletedAt(ctx, key, time.Now().UTC())
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) retrieveUserString(ctx context.Context, key string) (dbUserString schema.UserString, err error) {
	row := s.db.QueryRowContext(ctx, query.SelectUserString, key)

	err = row.Scan(&dbUserString.UserID, &dbUserString.Value, &dbUserString.DeletedAt)
	if err != nil {
		return schema.UserString{}, err
	}
	return dbUserString, nil
}

func (s *Storage) updateStringDeletedAt(ctx context.Context, key string, t time.Time) error {
	row := s.db.QueryRowContext(ctx, query.UpdateStringDeletedAt, key, t)
	return row.Err()
}

// RetrieveUserString is the sql implementation
func (s *Storage) RetrieveUserString(ctx context.Context, key string) (storage.UserString, error) {
	dus, err := s.retrieveUserString(ctx, key)
	if err != nil {
		return storage.UserString{}, err
	}

	return storage.UserString{
		UserID:  dus.UserID,
		Key:     key,
		Value:   dus.Value,
		Deleted: dus.IsDeleted(),
	}, nil
}

// TearDown provide closes and drops the underlying sql DB.
func (s *Storage) TearDown() error {
	err := s.db.Close()
	if err != nil {
		return err
	}

	err = s.drop()
	if err != nil {
		return err
	}
	return nil
}

func (s *Storage) drop() error {
	defaultDB, err := connectDB(string(s.DriverName), s.DefaultDSN)
	if err != nil {
		return err
	}
	defer func() { err = defaultDB.Close(); slog.Error(err.Error()) }()

	s.db.Close()
	if err != nil {
		return err
	}

	ctx := context.Background()
	q := fmt.Sprintf("DROP DATABASE %s;", s.DBName)
	_, err = defaultDB.ExecContext(ctx, q)
	if err != nil {
		return err
	}

	return nil
}

func createDB(dbCfg db.Config) error {
	defaultDB, err := connectDB(string(dbCfg.DriverName), dbCfg.DefaultDSN)
	if err != nil {
		return err
	}

	ctx := context.Background()
	q := fmt.Sprintf("CREATE DATABASE %s;", dbCfg.DBName)
	_, err = defaultDB.ExecContext(ctx, q)
	if err != nil {
		return err
	}

	return nil
}

func connectDB(driverName string, dsn string) (*sql.DB, error) {
	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

type Execer interface {
	ExecContext(ctx context.Context, q string) (sql.Result, error)
}
