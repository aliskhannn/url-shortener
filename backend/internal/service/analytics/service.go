package analytics

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/wb-go/wbf/retry"
	"github.com/wb-go/wbf/zlog"

	"github.com/aliskhannn/url-shortener/internal/model"
)

// analyticsRepository defines the interface for link analytics persistence operations.
type analyticsRepository interface {
	SaveAnalytics(ctx context.Context, event model.Analytics) (uuid.UUID, error)
	CountClicks(ctx context.Context, alias string) (int, error)
	GetClicksByDay(ctx context.Context, alias string) (map[string]int, error)
	GetClicksByUserAgent(ctx context.Context, alias string) (map[string]int, error)
}

// cache defines the interface for caching link analytics.
type cache interface {
	SetWithRetry(ctx context.Context, strategy retry.Strategy, key string, value interface{}) error
	GetWithRetry(ctx context.Context, strategy retry.Strategy, key string) (string, error)
}

// The Service provides methods for creating and retrieving link analytics.
type Service struct {
	repo  analyticsRepository
	cache cache
}

// NewService creates a new Service instance with repository and cache.
func NewService(repo analyticsRepository, cache cache) *Service {
	return &Service{repo: repo, cache: cache}
}

type SummaryOfAnalytics struct {
	Alias       string         `json:"alias"`
	TotalClicks int            `json:"total_clicks"`
	Daily       map[string]int `json:"daily"`      // clicks per day
	UserAgent   map[string]int `json:"user_agent"` // clicks per User_Agent
}

// SaveAnalytics save a link analytics and caches them.
func (s *Service) SaveAnalytics(ctx context.Context, strategy retry.Strategy, event model.Analytics) (uuid.UUID, error) {
	id, err := s.repo.SaveAnalytics(ctx, event)
	if err != nil {
		return uuid.Nil, fmt.Errorf("save analytics: %w", err)
	}

	event.ID = id

	go func() {
		summary, err := s.GetAnalyticsSummary(ctx, strategy, event.Alias)
		if err != nil {
			zlog.Logger.Error().Err(err).Str("alias", event.Alias).Msg("failed to get analytics summary for cache")
			return
		}

		if b, err := json.Marshal(summary); err == nil {
			if err := s.cache.SetWithRetry(ctx, strategy, "analytics:"+event.Alias, string(b)); err != nil {
				zlog.Logger.Error().Err(err).Str("alias", event.Alias).Msg("failed to cache aggregated analytics")
			}
		}

		zlog.Logger.Info().Str("alias", event.Alias).Msg("saved analytics")
	}()

	return id, nil
}

// GetAnalyticsSummary retrieves aggregated analytics for a short link.
func (s *Service) GetAnalyticsSummary(ctx context.Context, strategy retry.Strategy, alias string) (*SummaryOfAnalytics, error) {
	// Check cache first.
	if str, err := s.cache.GetWithRetry(ctx, strategy, "analytics:"+alias); err != nil {
		var summary SummaryOfAnalytics
		if err := json.Unmarshal([]byte(str), &summary); err == nil {
			return &summary, nil // cache hit
		}
	}

	total, err := s.repo.CountClicks(ctx, alias)
	if err != nil {
		return nil, fmt.Errorf("count clicks: %w", err)
	}

	daily, err := s.repo.GetClicksByDay(ctx, alias)
	if err != nil {
		return nil, fmt.Errorf("get clicks by day: %w", err)
	}

	ua, err := s.repo.GetClicksByUserAgent(ctx, alias)
	if err != nil {
		return nil, fmt.Errorf("get clicks by user agent: %w", err)
	}

	summary := &SummaryOfAnalytics{
		Alias:       alias,
		TotalClicks: total,
		Daily:       daily,
		UserAgent:   ua,
	}

	// Save to cache.
	if b, err := json.Marshal(summary); err == nil {
		if err := s.cache.SetWithRetry(ctx, strategy, "analytics:"+alias, string(b)); err != nil {
			zlog.Logger.Error().Err(err).Str("alias", alias).Msg("failed to cache summary analytics")
		}
	}

	return summary, nil
}
