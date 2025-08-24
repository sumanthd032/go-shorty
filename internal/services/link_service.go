package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"
	"github.com/sumanthd032/go-shorty/internal/repositories/db"
	"github.com/sumanthd032/go-shorty/pkg/utils"
)

var ErrAliasExists = errors.New("custom alias already exists")
var ErrLinkNotFound = errors.New("link not found")

// This struct will be the message we send to our background worker.
type ClickEvent struct {
	LinkID    int64  `json:"link_id"`
	Timestamp time.Time `json:"timestamp"`
	IPAddress string `json:"ip_address"`
	UserAgent string `json:"user_agent"`
	Referrer  string `json:"referrer"`
}

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

type CreateLinkParams struct {
	OriginalURL string
	CustomAlias string
	UserID      int64
}

func (s *LinkService) Create(ctx context.Context, params CreateLinkParams) (db.Link, error) {
	alias := params.CustomAlias
	if alias == "" {
		alias = utils.String()
	}

	createParams := db.CreateLinkParams{
		Alias:       alias,
		OriginalUrl: params.OriginalURL,
		UserID: pgtype.Int8{
			Int64: params.UserID,
			Valid: true,
		},
	}

	link, err := s.queries.CreateLink(ctx, createParams)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return db.Link{}, ErrAliasExists
		}
		return db.Link{}, fmt.Errorf("could not create link: %w", err)
	}

	return link, nil
}

// GetOriginalURLAndTrack finds a link's destination and publishes a click event.
func (s *LinkService) GetOriginalURLAndTrack(ctx context.Context, alias, ip, userAgent, referrer string) (string, error) {
	// 1. Try to get from cache first for speed.
	originalURL, err := s.cache.Get(ctx, alias).Result()
	if err == nil {
		// Cache Hit. We need the link ID to track the click.
		// Let's assume we also cache a small link object, not just the URL.
		// For now, we'll simplify and do a DB lookup in the background.
		go s.publishClickEvent(alias, ip, userAgent, referrer)
		return originalURL, nil
	}
	if err != redis.Nil {
		return "", fmt.Errorf("error fetching from cache: %w", err)
	}

	// 2. Cache Miss. Get from database.
	link, err := s.queries.GetLinkByAlias(ctx, alias)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrLinkNotFound
		}
		return "", fmt.Errorf("database error: %w", err)
	}

	// 3. Store in cache for next time.
	err = s.cache.Set(ctx, alias, link.OriginalUrl, 1*time.Hour).Err()
	if err != nil {
		log.Printf("Failed to cache link %s: %v", alias, err)
	}

	// 4. Publish click event in the background.
	event := ClickEvent{
		LinkID:    link.ID,
		Timestamp: time.Now(),
		IPAddress: ip,
		UserAgent: userAgent,
		Referrer:  referrer,
	}
	s.publishEvent(ctx, event)

	return link.OriginalUrl, nil
}

// publishClickEvent is a helper for cache hits where we don't have the link ID.
// In a production system, you'd want to cache the link ID as well.
func (s *LinkService) publishClickEvent(alias, ip, userAgent, referrer string) {
	link, err := s.queries.GetLinkByAlias(context.Background(), alias)
	if err != nil {
		log.Printf("Failed to get link for tracking alias %s: %v", alias, err)
		return
	}
	event := ClickEvent{
		LinkID:    link.ID,
		Timestamp: time.Now(),
		IPAddress: ip,
		UserAgent: userAgent,
		Referrer:  referrer,
	}
	s.publishEvent(context.Background(), event)
}

// publishEvent marshals the event to JSON and adds it to a Redis Stream.
func (s *LinkService) publishEvent(ctx context.Context, event ClickEvent) {
	eventJSON, err := json.Marshal(event)
	if err != nil {
		log.Printf("Failed to marshal click event: %v", err)
		return
	}

	err = s.cache.XAdd(ctx, &redis.XAddArgs{
		Stream: "clicks_stream",
		Values: map[string]interface{}{"event": eventJSON},
	}).Err()

	if err != nil {
		log.Printf("Failed to publish click event to Redis stream: %v", err)
	}
}