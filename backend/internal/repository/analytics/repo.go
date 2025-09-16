package analytics

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/wb-go/wbf/dbpg"

	"github.com/aliskhannn/url-shortener/internal/model"
)

var (
	ErrAnalyticsNotFound = errors.New("analytics not found")
)

// Repository provides methods to interact with analytics table.
type Repository struct {
	db *dbpg.DB
}

// NewRepository creates a new analytics repository.
func NewRepository(db *dbpg.DB) *Repository {
	return &Repository{db: db}
}

// SaveAnalytics inserts link analytics into the database and returns its ID.
func (r *Repository) SaveAnalytics(ctx context.Context, event model.Analytics) (uuid.UUID, error) {
	query := `
		INSERT INTO analytics (
		    link_id, user_agent, device_type, os, browser, ip_address
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id;
    `

	err := r.db.Master.QueryRowContext(
		ctx, query, event.LinkID, event.UserAgent, event.Device, event.OS, event.Browser, event.IP,
	).Scan(&event.ID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("insert analytics: %w", err)
	}

	return event.ID, nil
}

// GetAnalytics retrieves the link analytics by link id.
func (r *Repository) GetAnalytics(ctx context.Context, linkID uuid.UUID) (model.Analytics, error) {
	query := `
 		SELECT id, link_id, user_agent, device_type, os, browser, ip_address
		FROM analytics
		WHERE link_id = $1
 		ORDER BY created_at DESC;
    `

	var analytics model.Analytics
	err := r.db.Master.QueryRowContext(
		ctx, query, linkID,
	).Scan(
		&analytics.ID, &analytics.LinkID, &analytics.UserAgent,
		&analytics.Device, &analytics.OS, &analytics.Browser, &analytics.IP,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Analytics{}, ErrAnalyticsNotFound
		}

		return model.Analytics{}, fmt.Errorf("get analytics: %w", err)
	}

	return analytics, nil
}
