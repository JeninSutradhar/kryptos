package entry

import (
	"time"

	"github.com/google/uuid"
)

// PasswordEntry represents a single password entry.
type PasswordEntry struct {
	ID        string    `json:"id"` // Unique ID for each entry (UUID)
	Title     string    `json:"title"`
	Username  string    `json:"username"`
	Password  string    `json:"password"` // Encrypted!
	URL       string    `json:"url,omitempty"`
	Notes     string    `json:"notes,omitempty"`
	Tags      []string  `json:"tags,omitempty"` // New field for tags
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// NewPasswordEntry creates a new PasswordEntry with a UUID.
func NewPasswordEntry(title, username, password, url, notes string, tags []string) PasswordEntry {
	return PasswordEntry{
		ID:        uuid.New().String(),
		Title:     title,
		Username:  username,
		Password:  password,
		URL:       url,
		Notes:     notes,
		Tags:      tags,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}
