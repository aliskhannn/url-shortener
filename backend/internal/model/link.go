package model

import (
	"time"

	"github.com/google/uuid"
)

// Link represents a shortened URL entry.
type Link struct {
	ID        uuid.UUID `json:"id"`         // unique identifier
	URL       string    `json:"url"`        // original url
	Alias     string    `json:"alias"`      // short alias
	CreatedAt time.Time `json:"created_at"` // creation timestamp
}
