# AS-IP Server

HTTP API for IP, ASN, and Autonomous System geolocation lookups backed by PostgreSQL.

## Tech stack

- Go 1.25 + Gin
- PostgreSQL 16+ (via `github.com/jackc/pgx/v5`)
- Docker Compose for local deployment

## How to run

```bash
cp .env.example .env
.\.armin\docker-scripts\run-on-docker-local.ps1
```

Or: `docker compose up --build` (requires `docker network create asip-net` first).

API base URL: `http://localhost:3000/api/v1`

For local development without Docker:

```bash
psql -U postgres -c "CREATE DATABASE as_ip;"
psql -U postgres -d as_ip -f db/init.sql
go run ./cmd/api
```

## Entry points

| Command | Purpose |
|---------|---------|
| `go run ./cmd/api` | HTTP API server |
| `go run ./cmd/asip status` | Show last sync + request counts |
| `go run ./cmd/asip sync` | One-off data sync |
| `go run ./cmd/sync` | One-off data sync (standalone) |
