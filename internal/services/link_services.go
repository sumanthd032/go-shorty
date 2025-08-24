// File: internal/services/link_service.go

package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sumanthd032/go-shorty/internal/repositories/db"
	"github.com/sumanthd032/go-shorty/pkg/utils"
)

// Define a custom error for when a custom alias already exists.
var ErrAliasExists = errors.New("custom alias already exists")

// LinkService provides the business logic for link operations.
type LinkService struct {
	queries *db.Queries
}

// NewLinkService creates a new LinkService.
func NewLinkService(queries *db.Queries) *LinkService {
	return &LinkService{queries: queries}
}

// CreateLinkParams defines the parameters for creating a new link.
type CreateLinkParams struct {
	OriginalURL  string
	CustomAlias  string
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