#!/bin/sh

# This script runs database migrations and then starts the main server.

# Run migrations
/goose -dir "migrations" postgres "$DATABASE_DSN" up

# Start the server
/server
