# --- Stage 1: Build ---
# Your correct Go version
FROM golang:1.24.5-alpine AS builder

WORKDIR /app

# Install the goose binary inside the builder stage
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -o /worker ./cmd/worker


# --- Stage 2: Final Image ---
FROM gcr.io/distroless/static-debian12

# Copy our application binaries
COPY --from=builder /server /server
COPY --from=builder /worker /worker

# Copy the goose binary so we can run migrations
COPY --from=builder /go/bin/goose /goose

# Copy the necessary files and directories
COPY static /static
COPY config.yaml /config.yaml
COPY migrations /migrations