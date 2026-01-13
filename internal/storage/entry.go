package storage

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kitsnail/gpasswd/internal/crypto"
	"github.com/kitsnail/gpasswd/internal/models"
)

// EntryData represents the encrypted data stored in the database
type EntryData struct {
	Username string   `json:"username"`
	Password string   `json:"password"`
	URL      string   `json:"url"`
	Notes    string   `json:"notes"`
	Tags     []string `json:"tags"`
}

// CreateEntry encrypts and stores a new password entry in the database
// Assigns a new UUID, encrypts sensitive data, and stores with encryption metadata
func (db *DB) CreateEntry(entry *models.Entry, key []byte) error {
	// Validate input
	if entry == nil {
		return errors.New("entry cannot be nil")
	}
	if entry.Name == "" {
		return errors.New("entry name cannot be empty")
	}
	if entry.Password == "" {
		return errors.New("entry password cannot be empty")
	}
	if key == nil || len(key) != 32 {
		return errors.New("encryption key must be 32 bytes")
	}

	// Assign new ID if not set
	if entry.ID == "" {
		entry.ID = uuid.New().String()
	}

	// Set timestamps
	now := time.Now()
	entry.CreatedAt = now
	entry.UpdatedAt = now

	// Set default category if empty
	if entry.Category == "" {
		entry.Category = "general"
	}

	// Prepare data for encryption
	data := EntryData{
		Username: entry.Username,
		Password: entry.Password,
		URL:      entry.URL,
		Notes:    entry.Notes,
		Tags:     entry.Tags,
	}

	// Serialize to JSON
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal entry data: %w", err)
	}

	// Encrypt data
	encryptedData, err := crypto.Encrypt(dataJSON, key)
	if err != nil {
		return fmt.Errorf("failed to encrypt entry data: %w", err)
	}

	// Generate search text (name + category + tags + username + URL)
	searchText := entry.SearchText() + " " + entry.Username + " " + entry.URL
	searchTextBytes := []byte(searchText)

	// Encrypt search text
	encryptedSearch, err := crypto.Encrypt(searchTextBytes, key)
	if err != nil {
		return fmt.Errorf("failed to encrypt search text: %w", err)
	}

	// Extract nonces (first 12 bytes of each ciphertext)
	dataNonce := encryptedData[:12]
	searchNonce := encryptedSearch[:12]

	// Insert into database
	query := `
		INSERT INTO entries (
			id, name, category, encrypted_data, encrypted_search,
			created_at, updated_at, encryption_nonce, search_nonce
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = db.Exec(query,
		entry.ID, entry.Name, entry.Category,
		encryptedData, encryptedSearch,
		entry.CreatedAt, entry.UpdatedAt,
		dataNonce, searchNonce,
	)
	if err != nil {
		return fmt.Errorf("failed to insert entry: %w", err)
	}

	return nil
}

// GetEntry retrieves and decrypts a password entry by ID
func (db *DB) GetEntry(id string, key []byte) (*models.Entry, error) {
	// Validate input
	if id == "" {
		return nil, errors.New("entry ID cannot be empty")
	}
	if key == nil || len(key) != 32 {
		return nil, errors.New("encryption key must be 32 bytes")
	}

	query := `
		SELECT id, name, category, encrypted_data,
		       created_at, updated_at
		FROM entries
		WHERE id = ?
	`

	var entry models.Entry
	var encryptedData []byte

	err := db.QueryRow(query, id).Scan(
		&entry.ID, &entry.Name, &entry.Category, &encryptedData,
		&entry.CreatedAt, &entry.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("entry with ID %s not found", id)
		}
		return nil, fmt.Errorf("failed to query entry: %w", err)
	}

	// Decrypt data
	decryptedData, err := crypto.Decrypt(encryptedData, key)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt entry data: %w", err)
	}

	// Unmarshal JSON
	var data EntryData
	err = json.Unmarshal(decryptedData, &data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal entry data: %w", err)
	}

	// Populate entry fields
	entry.Username = data.Username
	entry.Password = data.Password
	entry.URL = data.URL
	entry.Notes = data.Notes
	entry.Tags = data.Tags

	return &entry, nil
}

// GetEntryByName retrieves and decrypts a password entry by name
func (db *DB) GetEntryByName(name string, key []byte) (*models.Entry, error) {
	// Validate input
	if name == "" {
		return nil, errors.New("entry name cannot be empty")
	}

	// Get ID by name first
	var id string
	query := "SELECT id FROM entries WHERE name = ?"
	err := db.QueryRow(query, name).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("entry with name %s not found", name)
		}
		return nil, fmt.Errorf("failed to query entry by name: %w", err)
	}

	// Use GetEntry to retrieve and decrypt
	return db.GetEntry(id, key)
}

// ListEntries returns a list of all entries (without decrypting passwords)
// This is used for displaying entry lists in the CLI
func (db *DB) ListEntries() ([]*models.Entry, error) {
	query := `
		SELECT id, name, category, created_at, updated_at
		FROM entries
		ORDER BY name ASC
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query entries: %w", err)
	}
	defer rows.Close()

	var entries []*models.Entry
	for rows.Next() {
		var entry models.Entry
		err := rows.Scan(
			&entry.ID, &entry.Name, &entry.Category,
			&entry.CreatedAt, &entry.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan entry: %w", err)
		}
		entries = append(entries, &entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating entries: %w", err)
	}

	return entries, nil
}

// ListEntriesByCategory returns entries filtered by category
func (db *DB) ListEntriesByCategory(category string) ([]*models.Entry, error) {
	query := `
		SELECT id, name, category, created_at, updated_at
		FROM entries
		WHERE category = ?
		ORDER BY name ASC
	`

	rows, err := db.Query(query, category)
	if err != nil {
		return nil, fmt.Errorf("failed to query entries by category: %w", err)
	}
	defer rows.Close()

	var entries []*models.Entry
	for rows.Next() {
		var entry models.Entry
		err := rows.Scan(
			&entry.ID, &entry.Name, &entry.Category,
			&entry.CreatedAt, &entry.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan entry: %w", err)
		}
		entries = append(entries, &entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating entries: %w", err)
	}

	return entries, nil
}

// UpdateEntry updates an existing entry with new encrypted data
func (db *DB) UpdateEntry(entry *models.Entry, key []byte) error {
	// Validate input
	if entry == nil {
		return errors.New("entry cannot be nil")
	}
	if entry.ID == "" {
		return errors.New("entry ID cannot be empty")
	}
	if entry.Name == "" {
		return errors.New("entry name cannot be empty")
	}
	if entry.Password == "" {
		return errors.New("entry password cannot be empty")
	}
	if key == nil || len(key) != 32 {
		return errors.New("encryption key must be 32 bytes")
	}

	// Update timestamp
	entry.UpdatedAt = time.Now()

	// Set default category if empty
	if entry.Category == "" {
		entry.Category = "general"
	}

	// Prepare data for encryption
	data := EntryData{
		Username: entry.Username,
		Password: entry.Password,
		URL:      entry.URL,
		Notes:    entry.Notes,
		Tags:     entry.Tags,
	}

	// Serialize to JSON
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal entry data: %w", err)
	}

	// Encrypt data
	encryptedData, err := crypto.Encrypt(dataJSON, key)
	if err != nil {
		return fmt.Errorf("failed to encrypt entry data: %w", err)
	}

	// Generate and encrypt search text
	searchText := entry.SearchText() + " " + entry.Username + " " + entry.URL
	searchTextBytes := []byte(searchText)
	encryptedSearch, err := crypto.Encrypt(searchTextBytes, key)
	if err != nil {
		return fmt.Errorf("failed to encrypt search text: %w", err)
	}

	// Extract nonces
	dataNonce := encryptedData[:12]
	searchNonce := encryptedSearch[:12]

	// Update database
	query := `
		UPDATE entries
		SET name = ?, category = ?, encrypted_data = ?, encrypted_search = ?,
		    updated_at = ?, encryption_nonce = ?, search_nonce = ?
		WHERE id = ?
	`

	result, err := db.Exec(query,
		entry.Name, entry.Category, encryptedData, encryptedSearch,
		entry.UpdatedAt, dataNonce, searchNonce, entry.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update entry: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("entry with ID %s not found", entry.ID)
	}

	return nil
}

// DeleteEntry removes an entry from the database
func (db *DB) DeleteEntry(id string) error {
	// Validate input
	if id == "" {
		return errors.New("entry ID cannot be empty")
	}

	query := "DELETE FROM entries WHERE id = ?"
	result, err := db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete entry: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("entry with ID %s not found", id)
	}

	return nil
}

// CountEntries returns the total number of entries
func (db *DB) CountEntries() (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM entries"
	err := db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count entries: %w", err)
	}
	return count, nil
}
