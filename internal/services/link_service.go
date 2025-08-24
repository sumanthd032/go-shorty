// File: internal/services/link_service.go

package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/redis/go-redis/v9"
	"github.com/sumanthd032/go-shorty/internal/repositories/db"
	"github.com/sumanthd032/go-shorty/pkg/utils"
	"github.com/jackc/pgx/v5/pgtype"
)

var ErrAliasExists = errors.New("custom alias already exists")
var ErrLinkNotFound = errors.New("link not found")

type LinkService struct {
	queries *db.Queries
	cache   *redis.Client 
}

func NewLinkService(queries *db.Queries, cache *redis.Client) *LinkService {
	return &LinkService{
		queries: queries,
		cache:   cache,
	}
}

// CreateLinkParams defines the parameters for creating a new link.
type CreateLinkParams struct {
	OriginalURL  string
	CustomAlias  string
	UserID       int64
}

// Create handles the creation of a short link.
// It generates a random alias if a custom one is not provided.
func (s *LinkService) Create(ctx context.Context, params CreateLinkParams) (db.Link, error) {
	alias := params.CustomAlias
	if alias == "" {
		// If no custom alias is provided, generate a random one.
		alias = utils.String()
	}

	createParams := db.CreateLinkParams{
		Alias:       alias,
		OriginalUrl: params.OriginalURL,
		UserID:      pgtype.Int8{Int64: params.UserID, Valid: true},
	}

	link, err := s.queries.CreateLink(ctx, createParams)
	if err != nil {
		// Check if the error is a unique constraint violation (code 23505 in Postgres).
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return db.Link{}, ErrAliasExists
		}
		// For any other database error, return a generic error.
		return db.Link{}, fmt.Errorf("could not create link: %w", err)
	}

	return link, nil
}

func (s *LinkService) GetOriginalURL(ctx context.Context, alias string) (string, error) {
	// 1. Try to get the URL from the cache
	originalURL, err := s.cache.Get(ctx, alias).Result()
	if err == nil {
		// Cache Hit! Return the URL.
		return originalURL, nil
	}

	if err != redis.Nil {
		// An actual error occurred with Redis, not just a cache miss.
		return "", fmt.Errorf("error fetching from cache: %w", err)
	}

	// 2. Cache Miss. Get the link from the database.
	link, err := s.queries.GetLinkByAlias(ctx, alias)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrLinkNotFound
		}
		return "", fmt.Errorf("database error: %w", err)
	}

	// 3. Store the result in the cache for next time.
	// We'll set an expiration of 1 hour for this example.
	err = s.cache.Set(ctx, alias, link.OriginalUrl, 1*time.Hour).Err()
	if err != nil {
		// If caching fails, we still proceed but should log the error.
		// For simplicity, we'll just return the URL. In a real app, you'd log this.
	}

	return link.OriginalUrl, nil
}