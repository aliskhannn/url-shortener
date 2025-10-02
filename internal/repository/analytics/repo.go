package analytics

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/wb-go/wbf/dbpg"

	"github.com/aliskhannn/url-shortener/internal/model"
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
		    alias, user_agent, device_type, os, browser, ip_address
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id;
    `

	err := r.db.QueryRowContext(
		ctx, query, event.Alias, event.UserAgent, event.Device, event.OS, event.Browser, event.IP,
	).Scan(&event.ID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("insert analytics: %w", err)
	}

	return event.ID, nil
}

// CountClicks returns the total number of clicks for a short link alias.
func (r *Repository) CountClicks(ctx context.Context, alias string) (int, error) {
	var count int

	query := `SELECT COUNT(*) FROM analytics WHERE alias = $1;`

	err := r.db.QueryRowContext(ctx, query, alias).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count clicks: %w", err)
	}

	return count, nil
}

// GetClicksByDay returns number of clicks grouped by day.
func (r *Repository) GetClicksByDay(ctx context.Context, alias string) (map[string]int, error) {
	query := `
		SELECT TO_CHAR(created_at, 'YYYY-MM-DD') AS day, COUNT(*)
		FROM analytics
		WHERE alias = $1
		GROUP BY day
		ORDER BY day DESC;
    `

	rows, err := r.db.QueryContext(ctx, query, alias)
	if err != nil {
		return nil, fmt.Errorf("query clicks by day: %w", err)
	}
	defer rows.Close()

	result := make(map[string]int)
	for rows.Next() {
		var day string
		var count int

		if err := rows.Scan(&day, &count); err != nil {
			return nil, fmt.Errorf("scan clicks by day: %w", err)
		}

		result[day] = count
	}

	return result, nil
}

// GetClicksByUserAgent returns number of clicks grouped by User-Agent.
func (r *Repository) GetClicksByUserAgent(ctx context.Context, alias string) (map[string]int, error) {
	query := `
		SELECT user_agent, COUNT(*) 
		FROM analytics
		WHERE alias = $1
		GROUP BY user_agent
		ORDER BY COUNT(*) DESC;
	`

	rows, err := r.db.QueryContext(ctx, query, alias)
	if err != nil {
		return nil, fmt.Errorf("query clicks by user-agent: %w", err)
	}
	defer rows.Close()

	result := make(map[string]int)
	for rows.Next() {
		var ua string
		var count int

		if err := rows.Scan(&ua, &count); err != nil {
			return nil, fmt.Errorf("scan clicks by user-agent: %w", err)
		}

		result[ua] = count
	}

	return result, nil
}
