package link

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

// linkRepository defines the interface for link persistence operations.
type linkRepository interface {
	CreateLink(ctx context.Context, link model.Link) (uuid.UUID, error)
	GetLinkByAlias(ctx context.Context, alias string) (model.Link, error)
}

// cache defines the interface for caching links.
type cache interface {
	SetWithRetry(ctx context.Context, strategy retry.Strategy, key string, value interface{}) error
	GetWithRetry(ctx context.Context, strategy retry.Strategy, key string) (string, error)
}

// The Service provides methods for creating and retrieving links.
type Service struct {
	repo  linkRepository
	cache cache
}

// NewService creates a new Service instance with repository and cache.
func NewService(repo linkRepository, cache cache) *Service {
	return &Service{repo: repo, cache: cache}
}

// CreateLink creates a new link and caches it.
func (s *Service) CreateLink(ctx context.Context, strategy retry.Strategy, link model.Link) (uuid.UUID, error) {
	id, err := s.repo.CreateLink(ctx, link)
	if err != nil {
		return uuid.Nil, fmt.Errorf("create link: %w", err)
	}

	link.ID = id

	// Marshal link into JSON before caching.
	b, err := json.Marshal(link)
	if err != nil {
		return uuid.Nil, fmt.Errorf("marshal link: %w", err)
	}

	// Cache the link.
	err = s.cache.SetWithRetry(ctx, strategy, link.Alias, string(b))
	if err != nil {
		zlog.Logger.Error().
			Err(err).
			Str("alias", link.Alias).
			Msg("failed to cache link")
	}

	return id, nil
}

// GetLinkByAlias retrieves a shortened link by its alias.
// It first tries to get the link from cache. If the cache misses,
// it fetches the link from the repository and updates the cache.
func (s *Service) GetLinkByAlias(ctx context.Context, strategy retry.Strategy, alias string) (model.Link, error) {
	var link model.Link

	// Check cache first.
	str, err := s.cache.GetWithRetry(ctx, strategy, alias)
	if err == nil {
		// Unmarshal cached JSON into a link.
		err = json.Unmarshal([]byte(str), &link)
		if err != nil {
			return model.Link{}, fmt.Errorf("unmarshal link: %w", err)
		}

		return link, nil // cache hit
	}

	// If cache misses, fetch from repo and update cache.
	if errors.Is(err, redis.Nil) {
		link, err = s.repo.GetLinkByAlias(ctx, alias)
		if err != nil {
			return model.Link{}, fmt.Errorf("get link by alias: %w", err)
		}

		// Marshal link into JSON before caching.
		b, err := json.Marshal(link)
		if err != nil {
			return model.Link{}, fmt.Errorf("marshal link: %w", err)
		}

		// Cache the link.
		err = s.cache.SetWithRetry(ctx, strategy, link.Alias, string(b))
		if err != nil {
			zlog.Logger.Error().
				Err(err).
				Str("alias", link.Alias).
				Msg("failed to cache link")
		}
	}

	return link, nil
}
