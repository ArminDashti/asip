# AS-IP Server (Go + Gin)

HTTP API for IP, ASN, and Autonomous System geolocation lookups backed by PostgreSQL.

## Requirements

- Go 1.22+
- PostgreSQL 16+ (or Docker)

## Quick start with Docker

```bash
cd server
docker compose up --build
```

API base URL: `http://localhost:3000/api/v1`

## CORS

Browser UIs (e.g. asip-webui) call the API cross-origin. Configure allowed origins with:

```env
CORS_ALLOWED_ORIGINS=http://localhost:5173,https://asip.xaigrok.ir
```

Default when unset: `http://localhost:5173`. Implemented with [`gin-contrib/cors`](https://github.com/gin-contrib/cors) (`GET` + `OPTIONS`).

## Local development

1. Start PostgreSQL and apply the schema:

```bash
psql -U postgres -c "CREATE DATABASE as_ip;"
psql -U postgres -d as_ip -f db/init.sql
```

2. Copy environment variables:

```bash
cp .env.example .env
```

3. Run the API:

```bash
go run ./cmd/api
```

## Build

```bash
go build -o bin/as-ip-api ./cmd/api
go build -o bin/as-ip-sync ./cmd/sync
go build -o bin/asip ./cmd/asip
```

## CLI

```bash
asip status   # last sync + request counts (today / yesterday)
asip sync     # one-off data sync
```

## Daily data sync

The API embeds a scheduler that runs once per day (default 02:00 UTC). It:

1. `git pull` (or `git clone`) the three ipverse datasets into local paths
2. **DELETE** all existing rows from import tables
3. **INSERT** fresh data from `as-metadata`, `as-ip-blocks`, and `geo-ip-blocks`

Run a one-off sync manually:

```bash
asip sync
# or
go run ./cmd/asip sync
```

See [Endpoints.md](../Endpoints.md) for sync-related environment variables.

## API endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/health` | Health check |
| GET | `/metrics` | Prometheus metrics (request rate / latency) |
| GET | `/api/v1/ip/info` | IP lookup for caller (ASN, AS, country) |
| GET | `/api/v1/ip/info/:ip` | IP lookup (ASN, AS, country) |
| GET | `/api/v1/ip/whois` | IP whois (RDAP) for caller |
| GET | `/api/v1/ip/whois/:ip` | IP whois (RDAP) by address |
| GET | `/api/v1/http/headers` | Echo request HTTP headers + client IP |
| GET | `/api/v1/dns/lookup/*domain` | DNS lookup (A/AAAA/NS/MX/TXT/CNAME) + ASN/country for resolved IPs |
| GET | `/api/v1/asn/list` | List all ASN numbers and handles |
| GET | `/api/v1/asn/search/:asn` | ASN details by number or handle (case-insensitive) |
| GET | `/api/v1/asn/to-as/:asn` | Map ASN to AS summary |
| GET | `/api/v1/as/search/:as` | AS details by handle or number (case-insensitive) |
| GET | `/api/v1/country/list` | List all countries (ISO code + name) |
| GET | `/api/v1/country/search/:country` | ASN list by ISO country code |

Country path parameters accept **ISO 3166-1 alpha-2 codes only** (e.g. `US`). All string lookups are **case-insensitive**.

### Example

```bash
curl http://localhost:3000/api/v1/health
curl http://localhost:3000/api/v1/ip/info
curl http://localhost:3000/api/v1/ip/info/8.8.8.8
curl http://localhost:3000/api/v1/ip/whois/8.8.8.8
curl http://localhost:3000/api/v1/http/headers
curl http://localhost:3000/api/v1/dns/lookup/example.com
curl http://localhost:3000/api/v1/dns/lookup/https://example.com
curl http://localhost:3000/api/v1/asn/search/15169
curl http://localhost:3000/api/v1/as/search/google
curl http://localhost:3000/api/v1/country/search/us
```

## Project layout

```
server/
├── cmd/api/              # Application entrypoint
├── cmd/asip/             # CLI (status, sync)
├── db/init.sql           # PostgreSQL schema
├── db/db.md              # Table/column reference
├── internal/
│   ├── config/           # Environment configuration
│   ├── database/         # PostgreSQL pool
│   ├── dto/              # JSON response types
│   ├── handler/          # HTTP handlers
│   ├── mapper/           # Domain → DTO mapping
│   ├── model/            # Domain models
│   ├── repository/       # SQL data access
│   ├── router/           # Gin route registration
│   ├── service/          # Business logic
│   └── sync/             # Git pull + DB import + daily scheduler
├── cmd/sync/             # One-off sync CLI
├── docker-compose.yml
├── Dockerfile
└── go.mod
```

## Response compatibility

ASN/AS search endpoints return the same JSON shape as the previous NestJS service (`AsnResponseDto`): `asn`, `metadata`, `stats`, and `lastAnnounced`.

List and search endpoints for country wrap results in `{ "items": [...], "total": N }`.

## Environment variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `3000` | HTTP port |
| `DB_HOST` | `localhost` | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USERNAME` | `postgres` | Database user |
| `DB_PASSWORD` | `postgres` | Database password |
| `DB_NAME` | `as_ip` | Database name |
| `DB_SSLMODE` | `disable` | PostgreSQL SSL mode |
| `GIN_MODE` | `debug` | Gin mode (`debug` or `release`) |
| `SYNC_ENABLED` | `true` | Enable daily background sync |
| `SYNC_ON_STARTUP` | `false` | Run sync immediately when API starts |
| `SYNC_HOUR_UTC` | `2` | Hour (0–23) for daily sync in UTC |
| `REPO_AS_METADATA_PATH` | (see `.env.example`) | Local clone path for as-metadata |
| `REPO_AS_IP_BLOCKS_PATH` | (see `.env.example`) | Local clone path for as-ip-blocks |
| `REPO_GEO_IP_BLOCKS_PATH` | (see `.env.example`) | Local clone path for geo-ip-blocks |
| `REPO_AS_METADATA_URL` | `https://github.com/ipverse/as-metadata` | Git remote URL |
| `REPO_AS_IP_BLOCKS_URL` | `https://github.com/ipverse/as-ip-blocks` | Git remote URL |
| `REPO_GEO_IP_BLOCKS_URL` | `https://github.com/ipverse/geo-ip-blocks` | Git remote URL |
