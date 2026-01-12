package models

import "time"

// Entry represents a password entry in the vault
type Entry struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`     // e.g., "Gmail Work"
	Category  string    `json:"category"` // e.g., "email", "api-key", "website"
	Username  string    `json:"username"` // optional
	Password  string    `json:"password"` // sensitive field
	URL       string    `json:"url"`      // optional
	Notes     string    `json:"notes"`    // optional, encrypted
	Tags      []string  `json:"tags"`     // e.g., ["work", "google"]
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SearchText generates the plain-text search index for the entry
func (e *Entry) SearchText() string {
	searchable := e.Name + " " + e.Category
	for _, tag := range e.Tags {
		searchable += " " + tag
	}
	return searchable
}
