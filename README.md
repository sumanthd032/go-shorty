# Go-Shorty: A Robust URL Shortener

Go-Shorty is a full-featured, high-performance URL shortener built with Go and a modern, reliable tech stack. It provides a clean API, a fast server-rendered UI, and a scalable architecture designed for production use. This project serves as a comprehensive example of building a real-world web application, from initial design to containerized deployment.

## ✨ Features
- **Short Link Creation:** Generate random short links or create custom aliases.
- **User Accounts:** Secure user registration and session-based login. Users can only manage their own links.
- **Click Tracking & Analytics:** Asynchronously tracks every click, recording IP, User-Agent, and referrer. A dedicated analytics page displays total clicks per link.
- **High-Performance Redirects:** Uses Redis caching for millisecond-level redirect speeds on popular links.
- **Background Processing:** A dedicated worker process handles click ingestion, ensuring the user-facing application is never slowed down by analytics processing.
- **Modern UI:** A clean, responsive user interface built with Tailwind CSS and vanilla JavaScript.

## 🛠️ Tech Stack

| Category   | Technology / Library |
|------------|-----------------------|
| Backend    | Go 1.24+             |
| HTTP/API   | Chi (v5)             |
| Database   | PostgreSQL (16)      |
| DB Tools   | pgx, sqlc, Goose     |
| Cache/Queue| Redis (7)            |
| Frontend   | HTML, Tailwind CSS, Vanilla JavaScript |
| Auth       | Bcrypt, Gorilla Sessions |
| Config     | Viper                |
| DevOps     | Docker, Docker Compose, GitHub Actions |

## 🏗️ High-Level Architecture

```plaintext
                               |         User's          |
                               |   Browser / API Client  |
                               +-----------+-------------+
                                           |
                                           | (HTTPS)
                                           v
+-----------------------------------------------------------------------------------+
|                                     Your Server                                   |
| +----------------------------------+ +------------------------------------------+ |
| |        Gateway / HTTP Server     | |              Workers (Async)             | |
| | (Go with Chi Router)             | |                                          | |
| |                                  | | +-----------------+                      | |
| | Handles:                         | | |  Click Processor  |                      | |
| |  - UI Requests (HTML/HTMX)       | | +-----------------+                      | |
| |  - API Requests (JSON)           | | |  Link Health      |                      | |
| |  - Redirects (/short-alias)      | | +-----------------+                      | |
| +----------------+-----------------+ +----------+-------------------------------+ |
|                  |                              ^                                 |
|                  | (Function Calls)             | (Jobs)                          |
|                  v                              |                                 |
| +----------------+-----------------+ +----------+-------------------------------+ |
| |       Core Service Layer         | |            Message Queue / Cache           | |
| | (Business Logic in Go)           | |            (Redis)                         | |
| |                                  | |                                          | |
| |  - Alias Generation              | |  - Caches hot links for fast redirects   | |
| |  - Validation & Security         | |  - Queues click data for processing      | |
| |  - User Authentication           | |  - Handles rate limiting                 | |
| +----------------+-----------------+ +------------------------------------------+ |
|                  |                                                               |
|                  | (SQL Queries via sqlc)                                        |
|                  v                                                               |
| +----------------+-------------------------------------------------------------+ |
| |                         Data Store (PostgreSQL)                                | |
| |                                                                                | |
| |  - Stores users, links, API keys, click data, etc.                             | |
| +--------------------------------------------------------------------------------+ |
+-----------------------------------------------------------------------------------+
```

- **Go Server:** Serves UI, JSON API, and handles redirects by publishing click events.
- **Go Worker:** Listens for Redis stream events and stores click analytics in PostgreSQL.
- **Redis:** High-speed cache and message queue.
- **PostgreSQL:** Primary data store.

## 🚀 Getting Started

### Prerequisites
- Docker
- Docker Compose
- Go (for local development, optional if using Docker only)
- Goose CLI (`go install github.com/pressly/goose/v3/cmd/goose@latest`)

### Installation & Setup

Clone the repository:

```bash
git clone https://github.com/sumanthd032/go-shorty.git
cd go-shorty
```

Configure the application:

Edit `config.yaml` and set `session_key` to a long, random 32 or 64-character string.

Build and run with Docker Compose:

```bash
docker compose up --build
```

Run database migrations:

```bash
docker compose exec server /goose -dir "migrations" postgres "postgresql://user:password@postgres:5432/shortener?sslmode=disable" up
```

### Usage

- **Web Interface:** Open [http://localhost:8080](http://localhost:8080)
- **Logs:** Check logs in the terminal where `docker-compose up` is running.
- **Stop:** Press `Ctrl+C`. To remove data volumes:  
  ```bash
  docker compose down --volumes
  ```