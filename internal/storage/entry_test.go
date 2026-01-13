package storage

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/kitsnail/gpasswd/internal/crypto"
	"github.com/kitsnail/gpasswd/internal/models"
)

// Helper function to create test database with initialized crypto
func createTestDBWithKey(t *testing.T) (*DB, []byte, func()) {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "test_entries.db")
	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Generate salt and save
	salt, err := crypto.GenerateSalt()
	if err != nil {
		t.Fatalf("Failed to generate salt: %v", err)
	}
	err = db.SetSalt(salt)
	if err != nil {
		t.Fatalf("Failed to set salt: %v", err)
	}

	// Set Argon2 params
	params := crypto.DefaultArgon2Params()
	err = db.SetArgon2Params(params)
	if err != nil {
		t.Fatalf("Failed to set Argon2 params: %v", err)
	}

	// Derive encryption key from test password
	password := "test-master-password-123"
	key, err := crypto.DeriveKey(password, salt, params)
	if err != nil {
		t.Fatalf("Failed to derive key: %v", err)
	}

	cleanup := func() {
		db.Close()
	}

	return db, key, cleanup
}

func TestCreateEntry(t *testing.T) {
	db, key, cleanup := createTestDBWithKey(t)
	defer cleanup()

	tests := []struct {
		name    string
		entry   *models.Entry
		wantErr bool
	}{
		{
			name: "valid - complete entry",
			entry: &models.Entry{
				Name:     "github.com",
				Category: "development",
				Username: "user@example.com",
				Password: "SecureP@ssw0rd123!",
				URL:      "https://github.com",
				Notes:    "My GitHub account",
				Tags:     []string{"work", "dev"},
			},
			wantErr: false,
		},
		{
			name: "valid - minimal entry",
			entry: &models.Entry{
				Name:     "minimal-entry",
				Password: "password123",
			},
			wantErr: false,
		},
		{
			name: "valid - with unicode",
			entry: &models.Entry{
				Name:     "中文账号",
				Username: "用户@测试.com",
				Password: "密码123!@#",
				Notes:    "包含中文的测试账号",
			},
			wantErr: false,
		},
		{
			name: "invalid - nil entry",
			entry: nil,
			wantErr: true,
		},
		{
			name: "invalid - empty name",
			entry: &models.Entry{
				Name:     "",
				Password: "password",
			},
			wantErr: true,
		},
		{
			name: "invalid - empty password",
			entry: &models.Entry{
				Name:     "test",
				Password: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := db.CreateEntry(tt.entry, key)

			if tt.wantErr {
				if err == nil {
					t.Error("CreateEntry() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("CreateEntry() unexpected error: %v", err)
				return
			}

			// Verify entry has ID assigned
			if tt.entry.ID == "" {
				t.Error("CreateEntry() did not assign ID to entry")
			}

			// Verify timestamps set
			if tt.entry.CreatedAt.IsZero() {
				t.Error("CreateEntry() did not set CreatedAt")
			}
			if tt.entry.UpdatedAt.IsZero() {
				t.Error("CreateEntry() did not set UpdatedAt")
			}
		})
	}
}

func TestCreateEntryDuplicateName(t *testing.T) {
	db, key, cleanup := createTestDBWithKey(t)
	defer cleanup()

	// Create first entry
	entry1 := &models.Entry{
		Name:     "duplicate-test",
		Password: "password1",
	}
	err := db.CreateEntry(entry1, key)
	if err != nil {
		t.Fatalf("CreateEntry() first entry error: %v", err)
	}

	// Try to create second entry with same name
	entry2 := &models.Entry{
		Name:     "duplicate-test",
		Password: "password2",
	}
	err = db.CreateEntry(entry2, key)
	if err == nil {
		t.Error("CreateEntry() should fail with duplicate name")
	}
}

func TestGetEntry(t *testing.T) {
	db, key, cleanup := createTestDBWithKey(t)
	defer cleanup()

	// Create test entry
	original := &models.Entry{
		Name:     "test-get",
		Category: "test-category",
		Username: "testuser@example.com",
		Password: "TestPassword123!",
		URL:      "https://test.example.com",
		Notes:    "Test notes with special chars: !@#$%^&*()",
		Tags:     []string{"tag1", "tag2", "tag3"},
	}
	err := db.CreateEntry(original, key)
	if err != nil {
		t.Fatalf("CreateEntry() setup error: %v", err)
	}

	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{
			name:    "valid - get existing entry",
			id:      original.ID,
			wantErr: false,
		},
		{
			name:    "invalid - non-existent ID",
			id:      "non-existent-id-12345",
			wantErr: true,
		},
		{
			name:    "invalid - empty ID",
			id:      "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry, err := db.GetEntry(tt.id, key)

			if tt.wantErr {
				if err == nil {
					t.Error("GetEntry() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("GetEntry() unexpected error: %v", err)
				return
			}

			// Verify all fields match original
			if entry.ID != original.ID {
				t.Errorf("GetEntry() ID = %s, want %s", entry.ID, original.ID)
			}
			if entry.Name != original.Name {
				t.Errorf("GetEntry() Name = %s, want %s", entry.Name, original.Name)
			}
			if entry.Category != original.Category {
				t.Errorf("GetEntry() Category = %s, want %s", entry.Category, original.Category)
			}
			if entry.Username != original.Username {
				t.Errorf("GetEntry() Username = %s, want %s", entry.Username, original.Username)
			}
			if entry.Password != original.Password {
				t.Errorf("GetEntry() Password = %s, want %s", entry.Password, original.Password)
			}
			if entry.URL != original.URL {
				t.Errorf("GetEntry() URL = %s, want %s", entry.URL, original.URL)
			}
			if entry.Notes != original.Notes {
				t.Errorf("GetEntry() Notes = %s, want %s", entry.Notes, original.Notes)
			}
			if len(entry.Tags) != len(original.Tags) {
				t.Errorf("GetEntry() Tags count = %d, want %d", len(entry.Tags), len(original.Tags))
			}
		})
	}
}

func TestGetEntryWithWrongKey(t *testing.T) {
	db, key, cleanup := createTestDBWithKey(t)
	defer cleanup()

	// Create entry with correct key
	entry := &models.Entry{
		Name:     "test-wrong-key",
		Password: "SecretPassword!",
	}
	err := db.CreateEntry(entry, key)
	if err != nil {
		t.Fatalf("CreateEntry() setup error: %v", err)
	}

	// Try to decrypt with wrong key
	wrongKey := make([]byte, 32)
	for i := range wrongKey {
		wrongKey[i] = byte(i + 1)
	}

	_, err = db.GetEntry(entry.ID, wrongKey)
	if err == nil {
		t.Error("GetEntry() with wrong key should fail (GCM authentication)")
	}
}

func TestListEntries(t *testing.T) {
	db, key, cleanup := createTestDBWithKey(t)
	defer cleanup()

	// Create multiple entries
	entries := []*models.Entry{
		{Name: "entry1", Category: "work", Password: "pass1"},
		{Name: "entry2", Category: "personal", Password: "pass2"},
		{Name: "entry3", Category: "work", Password: "pass3"},
	}

	for _, e := range entries {
		err := db.CreateEntry(e, key)
		if err != nil {
			t.Fatalf("CreateEntry() setup error: %v", err)
		}
	}

	// List all entries
	list, err := db.ListEntries()
	if err != nil {
		t.Errorf("ListEntries() error: %v", err)
	}

	if len(list) != 3 {
		t.Errorf("ListEntries() count = %d, want 3", len(list))
	}

	// Verify entry list contains essential fields
	for _, item := range list {
		if item.ID == "" {
			t.Error("ListEntries() entry missing ID")
		}
		if item.Name == "" {
			t.Error("ListEntries() entry missing Name")
		}
		if item.Category == "" {
			t.Error("ListEntries() entry missing Category")
		}
		// Password should NOT be included in list view
		if item.Password != "" {
			t.Error("ListEntries() should not include password in list")
		}
	}
}

func TestListEntriesByCategory(t *testing.T) {
	db, key, cleanup := createTestDBWithKey(t)
	defer cleanup()

	// Create entries with different categories
	entries := []*models.Entry{
		{Name: "work1", Category: "work", Password: "pass1"},
		{Name: "work2", Category: "work", Password: "pass2"},
		{Name: "personal1", Category: "personal", Password: "pass3"},
	}

	for _, e := range entries {
		err := db.CreateEntry(e, key)
		if err != nil {
			t.Fatalf("CreateEntry() setup error: %v", err)
		}
	}

	// List work category only
	workList, err := db.ListEntriesByCategory("work")
	if err != nil {
		t.Errorf("ListEntriesByCategory() error: %v", err)
	}

	if len(workList) != 2 {
		t.Errorf("ListEntriesByCategory('work') count = %d, want 2", len(workList))
	}

	// Verify all are work category
	for _, item := range workList {
		if item.Category != "work" {
			t.Errorf("ListEntriesByCategory('work') returned entry with category %s", item.Category)
		}
	}
}

func TestUpdateEntry(t *testing.T) {
	db, key, cleanup := createTestDBWithKey(t)
	defer cleanup()

	// Create original entry
	original := &models.Entry{
		Name:     "test-update",
		Category: "original",
		Username: "original@example.com",
		Password: "OriginalPass123",
		URL:      "https://original.com",
		Notes:    "Original notes",
		Tags:     []string{"original"},
	}
	err := db.CreateEntry(original, key)
	if err != nil {
		t.Fatalf("CreateEntry() setup error: %v", err)
	}

	originalCreatedAt := original.CreatedAt
	time.Sleep(100 * time.Millisecond) // Ensure timestamp difference

	// Update entry
	original.Category = "updated"
	original.Username = "updated@example.com"
	original.Password = "UpdatedPass456"
	original.URL = "https://updated.com"
	original.Notes = "Updated notes"
	original.Tags = []string{"updated", "modified"}

	err = db.UpdateEntry(original, key)
	if err != nil {
		t.Errorf("UpdateEntry() error: %v", err)
	}

	// Retrieve and verify updates
	retrieved, err := db.GetEntry(original.ID, key)
	if err != nil {
		t.Fatalf("GetEntry() error: %v", err)
	}

	if retrieved.Category != "updated" {
		t.Errorf("UpdateEntry() Category = %s, want 'updated'", retrieved.Category)
	}
	if retrieved.Username != "updated@example.com" {
		t.Errorf("UpdateEntry() Username = %s, want 'updated@example.com'", retrieved.Username)
	}
	if retrieved.Password != "UpdatedPass456" {
		t.Errorf("UpdateEntry() Password = %s, want 'UpdatedPass456'", retrieved.Password)
	}

	// Verify CreatedAt unchanged (ignoring location for DB round-trip)
	// Compare Unix seconds to avoid timezone/precision issues
	if retrieved.CreatedAt.Unix() != originalCreatedAt.Unix() {
		t.Errorf("UpdateEntry() should not change CreatedAt: got %v, want %v",
			retrieved.CreatedAt.Unix(), originalCreatedAt.Unix())
	}
	// UpdatedAt should be equal to or after CreatedAt
	if retrieved.UpdatedAt.Unix() < originalCreatedAt.Unix() {
		t.Errorf("UpdateEntry() UpdatedAt should not be before CreatedAt: UpdatedAt=%v, CreatedAt=%v",
			retrieved.UpdatedAt, originalCreatedAt)
	}
}

func TestUpdateEntryNonExistent(t *testing.T) {
	db, key, cleanup := createTestDBWithKey(t)
	defer cleanup()

	// Try to update non-existent entry
	nonExistent := &models.Entry{
		ID:       "non-existent-id",
		Name:     "test",
		Password: "test",
	}

	err := db.UpdateEntry(nonExistent, key)
	if err == nil {
		t.Error("UpdateEntry() should fail for non-existent entry")
	}
}

func TestDeleteEntry(t *testing.T) {
	db, key, cleanup := createTestDBWithKey(t)
	defer cleanup()

	// Create entry
	entry := &models.Entry{
		Name:     "test-delete",
		Password: "DeleteMe123",
	}
	err := db.CreateEntry(entry, key)
	if err != nil {
		t.Fatalf("CreateEntry() setup error: %v", err)
	}

	// Delete entry
	err = db.DeleteEntry(entry.ID)
	if err != nil {
		t.Errorf("DeleteEntry() error: %v", err)
	}

	// Verify entry no longer exists
	_, err = db.GetEntry(entry.ID, key)
	if err == nil {
		t.Error("GetEntry() after delete should fail")
	}
}

func TestDeleteEntryNonExistent(t *testing.T) {
	db, _, cleanup := createTestDBWithKey(t)
	defer cleanup()

	err := db.DeleteEntry("non-existent-id")
	if err == nil {
		t.Error("DeleteEntry() should fail for non-existent entry")
	}
}

func TestGetEntryByName(t *testing.T) {
	db, key, cleanup := createTestDBWithKey(t)
	defer cleanup()

	// Create entry
	original := &models.Entry{
		Name:     "test-by-name",
		Username: "user@example.com",
		Password: "TestPass123",
	}
	err := db.CreateEntry(original, key)
	if err != nil {
		t.Fatalf("CreateEntry() setup error: %v", err)
	}

	// Get by name
	retrieved, err := db.GetEntryByName("test-by-name", key)
	if err != nil {
		t.Errorf("GetEntryByName() error: %v", err)
	}

	if retrieved.ID != original.ID {
		t.Errorf("GetEntryByName() ID = %s, want %s", retrieved.ID, original.ID)
	}
	if retrieved.Password != original.Password {
		t.Errorf("GetEntryByName() Password mismatch")
	}

	// Try non-existent name
	_, err = db.GetEntryByName("non-existent-name", key)
	if err == nil {
		t.Error("GetEntryByName() should fail for non-existent name")
	}
}

func TestEntryEncryptionRoundTrip(t *testing.T) {
	db, key, cleanup := createTestDBWithKey(t)
	defer cleanup()

	// Test various special characters and data types
	testEntry := &models.Entry{
		Name:     "encryption-test",
		Username: "user+test@example.com",
		Password: "P@$$w0rd!@#$%^&*()_+-=[]{}|;:',.<>?/~`",
		URL:      "https://example.com/path?query=value&foo=bar#fragment",
		Notes:    "Line 1\nLine 2\n\tTabbed\n\"Quoted\"\n'Single quotes'\n\\Backslashes\\",
		Tags:     []string{"special", "chars", "!@#", "数字"},
	}

	err := db.CreateEntry(testEntry, key)
	if err != nil {
		t.Fatalf("CreateEntry() error: %v", err)
	}

	retrieved, err := db.GetEntry(testEntry.ID, key)
	if err != nil {
		t.Fatalf("GetEntry() error: %v", err)
	}

	// Verify exact match
	if retrieved.Password != testEntry.Password {
		t.Error("Encryption round-trip failed: password mismatch")
	}
	if retrieved.Notes != testEntry.Notes {
		t.Error("Encryption round-trip failed: notes mismatch")
	}
	if len(retrieved.Tags) != len(testEntry.Tags) {
		t.Error("Encryption round-trip failed: tags count mismatch")
	}
}

// Benchmark tests
func BenchmarkCreateEntry(b *testing.B) {
	dbPath := filepath.Join(b.TempDir(), "bench.db")
	db, _ := InitDB(dbPath)
	defer db.Close()

	salt, _ := crypto.GenerateSalt()
	db.SetSalt(salt)
	params := crypto.DefaultArgon2Params()
	db.SetArgon2Params(params)
	key, _ := crypto.DeriveKey("password", salt, params)

	entry := &models.Entry{
		Name:     "benchmark-entry",
		Username: "bench@example.com",
		Password: "BenchmarkPassword123!",
		URL:      "https://benchmark.com",
		Notes:    "Benchmark notes",
		Tags:     []string{"benchmark", "test"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		entry.Name = "bench-" + string(rune(i))
		db.CreateEntry(entry, key)
	}
}

func BenchmarkGetEntry(b *testing.B) {
	dbPath := filepath.Join(b.TempDir(), "bench_get.db")
	db, _ := InitDB(dbPath)
	defer db.Close()

	salt, _ := crypto.GenerateSalt()
	db.SetSalt(salt)
	params := crypto.DefaultArgon2Params()
	db.SetArgon2Params(params)
	key, _ := crypto.DeriveKey("password", salt, params)

	entry := &models.Entry{
		Name:     "benchmark-get",
		Password: "password",
	}
	db.CreateEntry(entry, key)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		db.GetEntry(entry.ID, key)
	}
}
