package analytics

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/wb-go/wbf/retry"
	"github.com/wb-go/wbf/zlog"

	"github.com/aliskhannn/url-shortener/internal/model"
)

// analyticsRepository defines the interface for link analytics persistence operations.
type analyticsRepository interface {
	SaveAnalytics(ctx context.Context, event model.Analytics) (uuid.UUID, error)
	GetAnalytics(ctx context.Context, alias string) (model.Analytics, error)
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

// SaveAnalytics save a link analytics and caches them.
func (s *Service) SaveAnalytics(ctx context.Context, strategy retry.Strategy, event model.Analytics) (uuid.UUID, error) {
	id, err := s.repo.SaveAnalytics(ctx, event)
	if err != nil {
		return uuid.Nil, fmt.Errorf("save analytics: %w", err)
	}

	event.ID = id

	// Marshal link analytics into JSON before caching.
	b, err := json.Marshal(event)
	if err != nil {
		return uuid.Nil, fmt.Errorf("marshal link: %w", err)
	}

	// Cache the link analytics.
	err = s.cache.SetWithRetry(ctx, strategy, event.Alias, string(b))
	if err != nil {
		zlog.Logger.Error().
			Err(err).
			Str("alias", event.Alias).
			Msg("failed to cache link analytics")
	}

	return id, nil
}

// GetAnalytics retrieves a link analytics by linkID.
// It first tries to get the link analytics from cache. If the cache misses,
// it fetches the link analytics from the repository and updates the cache.
func (s *Service) GetAnalytics(ctx context.Context, strategy retry.Strategy, alias string) (model.Analytics, error) {
	var event model.Analytics

	// Check cache first.
	str, err := s.cache.GetWithRetry(ctx, strategy, alias)
	if err == nil {
		// Unmarshal cached JSON into a link.
		err = json.Unmarshal([]byte(str), &event)
		if err != nil {
			return model.Analytics{}, fmt.Errorf("unmarshal link analytics: %w", err)
		}

		return event, nil // cache hit
	}

	// If cache misses, fetch from repo and update cache.
	if errors.Is(err, redis.Nil) {
		event, err = s.repo.GetAnalytics(ctx, alias)
		if err != nil {
			return model.Analytics{}, fmt.Errorf("get link by alias: %w", err)
		}

		// Marshal link into JSON before caching.
		b, err := json.Marshal(event)
		if err != nil {
			return model.Analytics{}, fmt.Errorf("marshal link: %w", err)
		}

		// Cache the link.
		err = s.cache.SetWithRetry(ctx, strategy, alias, string(b))
		if err != nil {
			zlog.Logger.Error().
				Err(err).
				Str("alias", alias).
				Msg("failed to cache link analytics")
		}
	}

	return event, nil
}
