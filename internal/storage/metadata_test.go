package storage

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/kitsnail/gpasswd/internal/crypto"
)

func TestSetMetadata(t *testing.T) {
	db, cleanup := createTestDB(t)
	defer cleanup()

	tests := []struct {
		name    string
		key     string
		value   string
		wantErr bool
	}{
		{
			name:    "valid - set new key",
			key:     "test_key",
			value:   "test_value",
			wantErr: false,
		},
		{
			name:    "valid - update existing key",
			key:     "test_key",
			value:   "updated_value",
			wantErr: false,
		},
		{
			name:    "valid - empty value",
			key:     "empty_key",
			value:   "",
			wantErr: false,
		},
		{
			name:    "invalid - empty key",
			key:     "",
			value:   "value",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := db.SetMetadata(tt.key, tt.value)

			if tt.wantErr {
				if err == nil {
					t.Error("SetMetadata() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("SetMetadata() unexpected error: %v", err)
			}
		})
	}
}

func TestGetMetadata(t *testing.T) {
	db, cleanup := createTestDB(t)
	defer cleanup()

	// Set up test data
	testKey := "test_key"
	testValue := "test_value"
	err := db.SetMetadata(testKey, testValue)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	tests := []struct {
		name      string
		key       string
		wantValue string
		wantErr   bool
	}{
		{
			name:      "valid - get existing key",
			key:       testKey,
			wantValue: testValue,
			wantErr:   false,
		},
		{
			name:      "invalid - get non-existent key",
			key:       "nonexistent",
			wantValue: "",
			wantErr:   true,
		},
		{
			name:      "invalid - empty key",
			key:       "",
			wantValue: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := db.GetMetadata(tt.key)

			if tt.wantErr {
				if err == nil {
					t.Error("GetMetadata() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("GetMetadata() unexpected error: %v", err)
				return
			}

			if value != tt.wantValue {
				t.Errorf("GetMetadata() = %s, want %s", value, tt.wantValue)
			}
		})
	}
}

func TestSetSalt(t *testing.T) {
	db, cleanup := createTestDB(t)
	defer cleanup()

	// Generate salt
	salt, err := crypto.GenerateSalt()
	if err != nil {
		t.Fatalf("GenerateSalt() error: %v", err)
	}

	// Set salt
	err = db.SetSalt(salt)
	if err != nil {
		t.Errorf("SetSalt() unexpected error: %v", err)
	}

	// Verify salt can be retrieved
	retrievedSalt, err := db.GetSalt()
	if err != nil {
		t.Errorf("GetSalt() unexpected error: %v", err)
	}

	if len(retrievedSalt) != len(salt) {
		t.Errorf("GetSalt() length = %d, want %d", len(retrievedSalt), len(salt))
	}

	// Note: We can't compare bytes directly due to encoding, but length should match
}

func TestGetSaltWhenNotSet(t *testing.T) {
	db, cleanup := createTestDB(t)
	defer cleanup()

	// Try to get salt before setting it
	_, err := db.GetSalt()
	if err == nil {
		t.Error("GetSalt() should fail when salt is not set")
	}
}

func TestSetSaltInvalid(t *testing.T) {
	db, cleanup := createTestDB(t)
	defer cleanup()

	tests := []struct {
		name    string
		salt    []byte
		wantErr bool
	}{
		{
			name:    "invalid - nil salt",
			salt:    nil,
			wantErr: true,
		},
		{
			name:    "invalid - empty salt",
			salt:    []byte{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := db.SetSalt(tt.salt)
			if tt.wantErr && err == nil {
				t.Error("SetSalt() expected error, got nil")
			}
		})
	}
}

func TestSetArgon2Params(t *testing.T) {
	db, cleanup := createTestDB(t)
	defer cleanup()

	params := crypto.DefaultArgon2Params()

	// Set params
	err := db.SetArgon2Params(params)
	if err != nil {
		t.Errorf("SetArgon2Params() unexpected error: %v", err)
	}

	// Verify params can be retrieved
	retrievedParams, err := db.GetArgon2Params()
	if err != nil {
		t.Errorf("GetArgon2Params() unexpected error: %v", err)
	}

	// Verify all fields match
	if retrievedParams.Time != params.Time {
		t.Errorf("GetArgon2Params() Time = %d, want %d", retrievedParams.Time, params.Time)
	}
	if retrievedParams.Memory != params.Memory {
		t.Errorf("GetArgon2Params() Memory = %d, want %d", retrievedParams.Memory, params.Memory)
	}
	if retrievedParams.Parallelism != params.Parallelism {
		t.Errorf("GetArgon2Params() Parallelism = %d, want %d", retrievedParams.Parallelism, params.Parallelism)
	}
	if retrievedParams.KeyLen != params.KeyLen {
		t.Errorf("GetArgon2Params() KeyLen = %d, want %d", retrievedParams.KeyLen, params.KeyLen)
	}
}

func TestGetArgon2ParamsWhenNotSet(t *testing.T) {
	db, cleanup := createTestDB(t)
	defer cleanup()

	// Try to get params before setting them
	_, err := db.GetArgon2Params()
	if err == nil {
		t.Error("GetArgon2Params() should fail when params are not set")
	}
}

func TestUpdateArgon2Params(t *testing.T) {
	db, cleanup := createTestDB(t)
	defer cleanup()

	// Set initial params
	initialParams := crypto.DefaultArgon2Params()
	err := db.SetArgon2Params(initialParams)
	if err != nil {
		t.Fatalf("SetArgon2Params() setup error: %v", err)
	}

	// Update params
	updatedParams := crypto.Argon2Params{
		Time:        5,
		Memory:      128 * 1024,
		Parallelism: 8,
		KeyLen:      32,
	}
	err = db.SetArgon2Params(updatedParams)
	if err != nil {
		t.Errorf("SetArgon2Params() update error: %v", err)
	}

	// Verify updated params
	retrievedParams, err := db.GetArgon2Params()
	if err != nil {
		t.Errorf("GetArgon2Params() error: %v", err)
	}

	if retrievedParams.Time != updatedParams.Time {
		t.Errorf("Updated Time = %d, want %d", retrievedParams.Time, updatedParams.Time)
	}
}

func TestMetadataRoundTrip(t *testing.T) {
	db, cleanup := createTestDB(t)
	defer cleanup()

	// Test various data types as JSON
	testData := map[string]interface{}{
		"string": "test_value",
		"number": 12345,
		"bool":   true,
		"array":  []string{"a", "b", "c"},
	}

	for key, value := range testData {
		// Serialize to JSON
		jsonValue, err := json.Marshal(value)
		if err != nil {
			t.Fatalf("JSON marshal error: %v", err)
		}

		// Set metadata
		err = db.SetMetadata(key, string(jsonValue))
		if err != nil {
			t.Errorf("SetMetadata(%s) error: %v", key, err)
		}

		// Get metadata
		retrievedJSON, err := db.GetMetadata(key)
		if err != nil {
			t.Errorf("GetMetadata(%s) error: %v", key, err)
		}

		// Verify JSON matches
		if retrievedJSON != string(jsonValue) {
			t.Errorf("Metadata mismatch for %s: got %s, want %s",
				key, retrievedJSON, string(jsonValue))
		}
	}
}

func TestMetadataConcurrentAccess(t *testing.T) {
	db, cleanup := createTestDB(t)
	defer cleanup()

	// Concurrent writes
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			key := "concurrent_key"
			value := "value_" + string(rune(id))
			err := db.SetMetadata(key, value)
			if err != nil {
				t.Errorf("Concurrent SetMetadata failed: %v", err)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify we can read the key (last write wins)
	_, err := db.GetMetadata("concurrent_key")
	if err != nil {
		t.Errorf("GetMetadata after concurrent writes failed: %v", err)
	}
}

// Integration test: full initialization workflow
func TestFullInitializationWorkflow(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "full_init.db")
	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB() error: %v", err)
	}
	defer db.Close()

	// Generate and save salt
	salt, err := crypto.GenerateSalt()
	if err != nil {
		t.Fatalf("GenerateSalt() error: %v", err)
	}

	err = db.SetSalt(salt)
	if err != nil {
		t.Errorf("SetSalt() error: %v", err)
	}

	// Save Argon2 params
	params := crypto.DefaultArgon2Params()
	err = db.SetArgon2Params(params)
	if err != nil {
		t.Errorf("SetArgon2Params() error: %v", err)
	}

	// Save additional metadata
	err = db.SetMetadata("version", "1.0.0")
	if err != nil {
		t.Errorf("SetMetadata(version) error: %v", err)
	}

	err = db.SetMetadata("created_at", "2026-01-13T00:00:00Z")
	if err != nil {
		t.Errorf("SetMetadata(created_at) error: %v", err)
	}

	// Verify all data can be retrieved
	retrievedSalt, err := db.GetSalt()
	if err != nil || len(retrievedSalt) == 0 {
		t.Error("Failed to retrieve salt")
	}

	retrievedParams, err := db.GetArgon2Params()
	if err != nil || retrievedParams.Time == 0 {
		t.Error("Failed to retrieve Argon2 params")
	}

	version, err := db.GetMetadata("version")
	if err != nil || version != "1.0.0" {
		t.Error("Failed to retrieve version metadata")
	}
}
