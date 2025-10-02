package model

import (
	"time"

	"github.com/google/uuid"
)

// Analytics represents a single visit to a shortened link.
type Analytics struct {
	ID        uuid.UUID `json:"id"`         // unique identifier
	Alias     string    `json:"alias"`      // short alias
	UserAgent string    `json:"user_agent"` // raw user agent string
	Device    string    `json:"device"`     // device type (desktop, mobile, tablet, bot)
	OS        string    `json:"os"`         // operating system
	Browser   string    `json:"browser"`    // browser name
	IP        string    `json:"ip"`         // client ip address
	CreatedAt time.Time `json:"created_at"` // timestamp of the visit
}
