# URL Shortener

A production-ready URL shortening service built with Go. Create short aliases for long URLs and redirect through them.

**Live Demo:** [url-shortener-t-production.up.railway.app](https://url-shortener-t-production.up.railway.app)

## Tech Stack

Go 1.24 • PostgreSQL • chi router • testify • Railway

## Features

- RESTful API with create/redirect/delete endpoints
- PostgreSQL storage with migrations
- Request validation and structured logging (slog)
- Basic authentication for protected routes
- Comprehensive unit tests with mocks
- Environment-based configuration (YAML + env vars)

## Implementation Details

This project started from a tutorial but includes several custom implementations:
- Migrated from SQLite to PostgreSQL with proper connection handling
- Implemented DELETE endpoint with error handling
- Wrote comprehensive unit tests covering edge cases and error scenarios
- Configured deployment pipeline for Railway
- Enhanced validation rules and error handling patterns

## API Endpoints

**Create short URL:**
```bash
curl -X POST http://localhost:8082/url \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com", "alias": "ex"}'
```

**Redirect:** `GET /{alias}` - redirects to original URL

**Delete:** `DELETE /url/{alias}` - removes short URL

## Local Setup

```bash
# Create database
createdb urlshortener

# Set configuration path
export CONFIG_PATH=./config/local.yaml

# Run migrations
migrate -path migrations -database "postgres://localhost:5432/urlshortener?sslmode=disable" up

# Start server
go run cmd/url-shortener/main.go
```

Server runs on `localhost:8082`

## Testing

```bash
go test ./...           # Run all tests
go test -cover ./...    # Run with coverage
go generate ./...       # Generate mocks
```

## Architecture

```
cmd/url-shortener/       - Application entry point
internal/
  ├── config/            - Configuration management
  ├── http-server/
  │   ├── handlers/      - HTTP handlers (save, redirect, delete)
  │   └── middleware/    - Logger and auth middleware
  ├── lib/               - Shared utilities
  └── storage/postgres/  - PostgreSQL implementation
migrations/              - Database migrations
config/                  - Environment configurations
```

## Configuration

Uses environment-based configuration with YAML fallback:

**Environment Variables:**
- `ENV` - Environment (local/prod)
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`
- `HTTP_USER`, `HTTP_PASSWORD` - Auth credentials
- `PORT` - Server port

## Deployment

Deployed on Railway with PostgreSQL service. Automatic deployments from main branch.
