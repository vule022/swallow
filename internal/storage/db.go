package storage

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// DB wraps the SQLite database connection.
type DB struct {
	db *sql.DB
}

// Open opens (or creates) the SQLite database at path and runs migrations.
func Open(path string) (*DB, error) {
	sqlDB, err := sql.Open("sqlite", path+"?_foreign_keys=on&_journal_mode=WAL")
	if err != nil {
		return nil, fmt.Errorf("storage: open db: %w", err)
	}
	sqlDB.SetMaxOpenConns(1)

	if err := migrate(sqlDB); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("storage: migrate: %w", err)
	}

	return &DB{db: sqlDB}, nil
}

// Close closes the database connection.
func (d *DB) Close() error {
	return d.db.Close()
}

// SQL returns the underlying *sql.DB for use by repositories.
func (d *DB) SQL() *sql.DB {
	return d.db
}
