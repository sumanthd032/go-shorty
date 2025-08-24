package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"
	"github.com/sumanthd032/go-shorty/internal/config"
	"github.com/sumanthd032/go-shorty/internal/repositories/db"
	"github.com/sumanthd032/go-shorty/internal/services" // <-- FIX: Corrected the import path
)

func main() {
	log.Println("Starting click processing worker...")
	ctx := context.Background()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	conn, err := pgx.Connect(ctx, cfg.Database.DSN)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer conn.Close(ctx)

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if _, err := rdb.Ping(ctx).Result(); err != nil {
		log.Fatalf("Unable to connect to Redis: %v", err)
	}

	queries := db.New(conn)
	streamName := "clicks_stream"
	groupName := "clicks_group"

	err = rdb.XGroupCreateMkStream(ctx, streamName, groupName, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		log.Printf("Error creating consumer group: %v", err)
	}

	log.Printf("Worker is listening for messages on stream '%s' in group '%s'", streamName, groupName)

	for {
		// FIX: Corrected the struct name from XRead-groupArgs to XReadGroupArgs
		streams, err := rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    groupName,
			Consumer: "consumer-1",
			Streams:  []string{streamName, ">"},
			Count:    1,
			Block:    0,
		}).Result()

		if err != nil {
			log.Printf("Error reading from stream: %v", err)
			continue
		}

		for _, stream := range streams {
			for _, message := range stream.Messages {
				eventJSON, ok := message.Values["event"].(string)
				if !ok {
					log.Println("Invalid message format: 'event' field is not a string")
					continue
				}

				var event services.ClickEvent
				if err := json.Unmarshal([]byte(eventJSON), &event); err != nil {
					log.Printf("Failed to unmarshal event: %v", err)
					continue
				}

				_, err := queries.CreateClick(ctx, db.CreateClickParams{
					LinkID:    event.LinkID,
					IpAddress: pgtype.Text{String: event.IPAddress, Valid: true},
					UserAgent: pgtype.Text{String: event.UserAgent, Valid: true},
					Referrer:  pgtype.Text{String: event.Referrer, Valid: true},
				})

				if err != nil {
					log.Printf("Failed to save click to database: %v", err)
				} else {
					log.Printf("Processed click for LinkID %d", event.LinkID)
					rdb.XAck(ctx, streamName, groupName, message.ID)
				}
			}
		}
	}
}