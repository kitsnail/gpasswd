package storage

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/kitsnail/gpasswd/internal/crypto"
)

// Metadata keys
const (
	MetadataKeySalt          = "salt"
	MetadataKeyArgon2Params  = "argon2_params"
	MetadataKeyVersion       = "version"
	MetadataKeyCreatedAt     = "created_at"
)

// SetMetadata stores a key-value pair in the metadata table
// If the key already exists, it will be updated (UPSERT)
func (db *DB) SetMetadata(key, value string) error {
	if key == "" {
		return errors.New("metadata key cannot be empty")
	}

	query := `
		INSERT INTO metadata (key, value) VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value
	`

	_, err := db.Exec(query, key, value)
	if err != nil {
		return fmt.Errorf("failed to set metadata %s: %w", key, err)
	}

	return nil
}

// GetMetadata retrieves a value from the metadata table
// Returns error if key doesn't exist
func (db *DB) GetMetadata(key string) (string, error) {
	if key == "" {
		return "", errors.New("metadata key cannot be empty")
	}

	var value string
	query := "SELECT value FROM metadata WHERE key = ?"

	err := db.QueryRow(query, key).Scan(&value)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("metadata key %s not found", key)
		}
		return "", fmt.Errorf("failed to get metadata %s: %w", key, err)
	}

	return value, nil
}

// SetSalt stores the salt used for key derivation
// Salt is base64-encoded for storage
func (db *DB) SetSalt(salt []byte) error {
	if salt == nil || len(salt) == 0 {
		return errors.New("salt cannot be nil or empty")
	}

	// Encode salt to base64 for storage
	encoded := base64.StdEncoding.EncodeToString(salt)

	return db.SetMetadata(MetadataKeySalt, encoded)
}

// GetSalt retrieves the salt used for key derivation
// Returns decoded binary salt
func (db *DB) GetSalt() ([]byte, error) {
	encoded, err := db.GetMetadata(MetadataKeySalt)
	if err != nil {
		return nil, fmt.Errorf("failed to get salt: %w", err)
	}

	// Decode from base64
	salt, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("failed to decode salt: %w", err)
	}

	return salt, nil
}

// SetArgon2Params stores the Argon2 parameters used for key derivation
// Parameters are stored as JSON
func (db *DB) SetArgon2Params(params crypto.Argon2Params) error {
	// Validate params
	if err := params.Validate(); err != nil {
		return fmt.Errorf("invalid Argon2 parameters: %w", err)
	}

	// Serialize to JSON
	jsonData, err := json.Marshal(params)
	if err != nil {
		return fmt.Errorf("failed to marshal Argon2 params: %w", err)
	}

	return db.SetMetadata(MetadataKeyArgon2Params, string(jsonData))
}

// GetArgon2Params retrieves the Argon2 parameters
// Returns error if not found or invalid
func (db *DB) GetArgon2Params() (crypto.Argon2Params, error) {
	jsonData, err := db.GetMetadata(MetadataKeyArgon2Params)
	if err != nil {
		return crypto.Argon2Params{}, fmt.Errorf("failed to get Argon2 params: %w", err)
	}

	var params crypto.Argon2Params
	err = json.Unmarshal([]byte(jsonData), &params)
	if err != nil {
		return crypto.Argon2Params{}, fmt.Errorf("failed to unmarshal Argon2 params: %w", err)
	}

	// Validate params
	if err := params.Validate(); err != nil {
		return crypto.Argon2Params{}, fmt.Errorf("invalid Argon2 parameters in database: %w", err)
	}

	return params, nil
}

// DeleteMetadata removes a key from the metadata table
func (db *DB) DeleteMetadata(key string) error {
	if key == "" {
		return errors.New("metadata key cannot be empty")
	}

	query := "DELETE FROM metadata WHERE key = ?"
	result, err := db.Exec(query, key)
	if err != nil {
		return fmt.Errorf("failed to delete metadata %s: %w", key, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("metadata key %s not found", key)
	}

	return nil
}

// ListMetadataKeys returns all metadata keys
func (db *DB) ListMetadataKeys() ([]string, error) {
	query := "SELECT key FROM metadata ORDER BY key"

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list metadata keys: %w", err)
	}
	defer rows.Close()

	var keys []string
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			return nil, fmt.Errorf("failed to scan metadata key: %w", err)
		}
		keys = append(keys, key)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating metadata keys: %w", err)
	}

	return keys, nil
}
