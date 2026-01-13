package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitDB(t *testing.T) {
	tests := []struct {
		name    string
		dbPath  string
		wantErr bool
	}{
		{
			name:    "valid - create new database",
			dbPath:  filepath.Join(t.TempDir(), "test.db"),
			wantErr: false,
		},
		{
			name:    "valid - existing database",
			dbPath:  filepath.Join(t.TempDir(), "existing.db"),
			wantErr: false,
		},
		{
			name:    "invalid - empty path",
			dbPath:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For existing database test, create it first
			if tt.name == "valid - existing database" {
				db, err := InitDB(tt.dbPath)
				if err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
				db.Close()
			}

			// Now test initialization
			db, err := InitDB(tt.dbPath)

			if tt.wantErr {
				if err == nil {
					t.Error("InitDB() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("InitDB() unexpected error: %v", err)
				return
			}

			if db == nil {
				t.Error("InitDB() returned nil database")
				return
			}

			// Verify database file exists
			if _, err := os.Stat(tt.dbPath); os.IsNotExist(err) {
				t.Errorf("InitDB() did not create database file at %s", tt.dbPath)
			}

			// Clean up
			db.Close()
		})
	}
}

func TestCreateSchema(t *testing.T) {
	// Create temporary database
	dbPath := filepath.Join(t.TempDir(), "schema_test.db")
	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB() error: %v", err)
	}
	defer db.Close()

	// Verify metadata table exists
	var tableName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='metadata'").Scan(&tableName)
	if err != nil {
		t.Errorf("metadata table not found: %v", err)
	}
	if tableName != "metadata" {
		t.Errorf("metadata table name = %s, want 'metadata'", tableName)
	}

	// Verify entries table exists
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='entries'").Scan(&tableName)
	if err != nil {
		t.Errorf("entries table not found: %v", err)
	}

	// Note: entries_fts table temporarily disabled (requires FTS5 support)
	// Will be re-enabled in future iteration
}

func TestMetadataTableSchema(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "metadata_schema_test.db")
	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB() error: %v", err)
	}
	defer db.Close()

	// Query metadata table schema
	rows, err := db.Query("PRAGMA table_info(metadata)")
	if err != nil {
		t.Fatalf("Failed to query metadata schema: %v", err)
	}
	defer rows.Close()

	expectedColumns := map[string]bool{
		"key":   false,
		"value": false,
	}

	for rows.Next() {
		var cid int
		var name, dataType string
		var notNull, pk int
		var dfltValue interface{}

		err := rows.Scan(&cid, &name, &dataType, &notNull, &dfltValue, &pk)
		if err != nil {
			t.Fatalf("Failed to scan column info: %v", err)
		}

		if _, exists := expectedColumns[name]; exists {
			expectedColumns[name] = true
		}
	}

	// Verify all expected columns exist
	for col, found := range expectedColumns {
		if !found {
			t.Errorf("metadata table missing column: %s", col)
		}
	}
}

func TestEntriesTableSchema(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "entries_schema_test.db")
	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB() error: %v", err)
	}
	defer db.Close()

	// Query entries table schema
	rows, err := db.Query("PRAGMA table_info(entries)")
	if err != nil {
		t.Fatalf("Failed to query entries schema: %v", err)
	}
	defer rows.Close()

	expectedColumns := map[string]bool{
		"id":                   false,
		"name":                 false,
		"category":             false,
		"encrypted_data":       false,
		"encrypted_search":     false,
		"created_at":           false,
		"updated_at":           false,
		"encryption_nonce":     false,
		"search_nonce":         false,
	}

	for rows.Next() {
		var cid int
		var name, dataType string
		var notNull, pk int
		var dfltValue interface{}

		err := rows.Scan(&cid, &name, &dataType, &notNull, &dfltValue, &pk)
		if err != nil {
			t.Fatalf("Failed to scan column info: %v", err)
		}

		if _, exists := expectedColumns[name]; exists {
			expectedColumns[name] = true
		}
	}

	// Verify all expected columns exist
	for col, found := range expectedColumns {
		if !found {
			t.Errorf("entries table missing column: %s", col)
		}
	}
}

func TestDatabaseClose(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "close_test.db")
	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB() error: %v", err)
	}

	// Close database
	err = db.Close()
	if err != nil {
		t.Errorf("Close() unexpected error: %v", err)
	}

	// Verify database is closed (query should fail)
	err = db.Ping()
	if err == nil {
		t.Error("Database should be closed, but Ping() succeeded")
	}
}

func TestDatabaseConcurrency(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "concurrent_test.db")
	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB() error: %v", err)
	}
	defer db.Close()

	// Test concurrent writes
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			_, err := db.Exec("INSERT INTO metadata (key, value) VALUES (?, ?)",
				"test_key_" + string(rune(id)), "test_value")
			if err != nil {
				t.Errorf("Concurrent write failed: %v", err)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestInitDBWithInvalidPath(t *testing.T) {
	// Try to create database in non-existent directory
	dbPath := "/nonexistent/directory/test.db"
	_, err := InitDB(dbPath)
	if err == nil {
		t.Error("InitDB() should fail with invalid path, got nil error")
	}
}

func TestForeignKeyConstraints(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "fk_test.db")
	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB() error: %v", err)
	}
	defer db.Close()

	// Check if foreign keys are enabled
	var fkEnabled int
	err = db.QueryRow("PRAGMA foreign_keys").Scan(&fkEnabled)
	if err != nil {
		t.Fatalf("Failed to query foreign_keys: %v", err)
	}

	if fkEnabled != 1 {
		t.Error("Foreign keys should be enabled")
	}
}

func TestWALMode(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "wal_test.db")
	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB() error: %v", err)
	}
	defer db.Close()

	// Check if WAL mode is enabled
	var journalMode string
	err = db.QueryRow("PRAGMA journal_mode").Scan(&journalMode)
	if err != nil {
		t.Fatalf("Failed to query journal_mode: %v", err)
	}

	if journalMode != "wal" {
		t.Errorf("Journal mode = %s, want 'wal'", journalMode)
	}
}

// Helper function to create test database
func createTestDB(t *testing.T) (*DB, func()) {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	cleanup := func() {
		db.Close()
	}

	return db, cleanup
}

// Benchmark tests
func BenchmarkInitDB(b *testing.B) {
	tempDir := b.TempDir()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dbPath := filepath.Join(tempDir, "bench_"+string(rune(i))+".db")
		db, err := InitDB(dbPath)
		if err != nil {
			b.Fatal(err)
		}
		db.Close()
	}
}
