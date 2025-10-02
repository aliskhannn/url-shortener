package link

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/wb-go/wbf/dbpg"

	"github.com/aliskhannn/url-shortener/internal/model"
)

var (
	ErrAliasNotFound = errors.New("link not found")
)

// Repository provides methods to interact with links table.
type Repository struct {
	db *dbpg.DB
}

// NewRepository creates a new link repository.
func NewRepository(db *dbpg.DB) *Repository {
	return &Repository{db: db}
}

// CreateLink inserts a new link into the database and returns its ID.
func (r *Repository) CreateLink(ctx context.Context, link model.Link) (model.Link, error) {
	query := `
		INSERT INTO links (url, alias)
		VALUES ($1, $2)
		RETURNING id, alias, created_at;
    `

	err := r.db.QueryRowContext(ctx, query, link.URL, link.Alias).Scan(&link.ID, &link.Alias, &link.CreatedAt)
	if err != nil {
		return model.Link{}, fmt.Errorf("insert link: %w", err)
	}

	return link, nil
}

// GetLinkByAlias retrieves the link by its alias.
func (r *Repository) GetLinkByAlias(ctx context.Context, alias string) (model.Link, error) {
	query := `
		SELECT id, url, alias
		FROM links
		WHERE alias = $1;
    `

	var link model.Link
	err := r.db.QueryRowContext(
		ctx, query, alias,
	).Scan(&link.ID, &link.URL, &link.Alias)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Link{}, ErrAliasNotFound
		}

		return model.Link{}, fmt.Errorf("get link by alias: %w", err)
	}

	return link, nil
}
