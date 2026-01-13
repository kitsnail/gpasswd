package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// DB wraps sql.DB with additional functionality for gpasswd
type DB struct {
	*sql.DB
	path string
}

// InitDB initializes and returns a new database connection
// Creates the database file if it doesn't exist
// Sets up the schema (tables, indexes, triggers)
// Configures SQLite for optimal performance and security
func InitDB(dbPath string) (*DB, error) {
	// Validate path
	if dbPath == "" {
		return nil, errors.New("database path cannot be empty")
	}

	// Ensure parent directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open database connection
	// Note: go-sqlite3 creates the file if it doesn't exist
	sqlDB, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxOpenConns(1) // SQLite works best with single connection
	sqlDB.SetMaxIdleConns(1)

	// Wrap in our DB type
	db := &DB{
		DB:   sqlDB,
		path: dbPath,
	}

	// Configure SQLite
	if err := db.configure(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to configure database: %w", err)
	}

	// Create schema
	if err := db.createSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create schema: %w", err)
	}

	return db, nil
}

// configure sets up SQLite pragmas for optimal performance and security
func (db *DB) configure() error {
	pragmas := []string{
		// Enable foreign key constraints
		"PRAGMA foreign_keys = ON",

		// Use Write-Ahead Logging for better concurrency
		"PRAGMA journal_mode = WAL",

		// Synchronous NORMAL is safe with WAL and much faster
		"PRAGMA synchronous = NORMAL",

		// Memory-mapped I/O for better performance
		"PRAGMA mmap_size = 30000000000", // 30GB

		// Increase cache size (negative value = KB)
		"PRAGMA cache_size = -64000", // 64MB

		// Use busy timeout to handle lock contention
		"PRAGMA busy_timeout = 5000", // 5 seconds
	}

	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			return fmt.Errorf("failed to execute pragma %s: %w", pragma, err)
		}
	}

	return nil
}

// createSchema creates all necessary tables and indexes
func (db *DB) createSchema() error {
	schema := `
	-- Metadata table for storing configuration and secrets
	-- Stores salt, Argon2 parameters, version info, etc.
	CREATE TABLE IF NOT EXISTS metadata (
		key TEXT PRIMARY KEY NOT NULL,
		value TEXT NOT NULL
	);

	-- Entries table for storing encrypted password entries
	CREATE TABLE IF NOT EXISTS entries (
		id TEXT PRIMARY KEY NOT NULL,
		name TEXT NOT NULL UNIQUE,
		category TEXT NOT NULL DEFAULT 'general',

		-- Encrypted data (JSON containing username, password, URL, notes, tags)
		encrypted_data BLOB NOT NULL,

		-- Encrypted search text for FTS (name + username + URL + category)
		encrypted_search BLOB NOT NULL,

		-- Timestamps
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

		-- Encryption metadata (nonces for GCM)
		encryption_nonce BLOB NOT NULL,
		search_nonce BLOB NOT NULL
	);

	-- Index for category filtering
	CREATE INDEX IF NOT EXISTS idx_entries_category ON entries(category);

	-- Index for timestamps (for sorting)
	CREATE INDEX IF NOT EXISTS idx_entries_created_at ON entries(created_at);
	CREATE INDEX IF NOT EXISTS idx_entries_updated_at ON entries(updated_at);

	-- Full-text search table (FTS5)
	-- This will store decrypted search text temporarily during search operations
	-- NOT persisted - populated on-demand during searches
	-- NOTE: Temporarily disabled - requires SQLite with FTS5 support
	-- Will be re-enabled in future iteration with proper SQLite build tags
	-- CREATE VIRTUAL TABLE IF NOT EXISTS entries_fts USING fts5(
	--	entry_id UNINDEXED,
	--	search_text,
	--	content='',
	--	tokenize='porter unicode61'
	-- );

	-- Trigger to update updated_at timestamp
	CREATE TRIGGER IF NOT EXISTS update_entries_timestamp
	AFTER UPDATE ON entries
	BEGIN
		UPDATE entries SET updated_at = CURRENT_TIMESTAMP
		WHERE id = NEW.id;
	END;
	`

	_, err := db.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
	}

	return nil
}

// Path returns the database file path
func (db *DB) Path() string {
	return db.path
}
