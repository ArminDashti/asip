# Database Module

PostgreSQL data layer for AS-IP geolocation lookups.

## Responsibility

- Open PostgreSQL connection pool via `pgx` stdlib driver
- Bootstrap schema from `db/init.sql` on first run
- Bulk insert helper for sync imports
- IPv4-to-integer conversion for range lookups

## Key files

| File | Description |
|------|-------------|
| `internal/database/postgres.go` | Connection pool, schema init, migration |
| `internal/database/bulk.go` | Batched INSERT with `$n` placeholders |
| `internal/database/iprange.go` | IPv4 string → integer conversion |
| `db/init.sql` | PostgreSQL DDL (14 tables) |
| `db/db.md` | Table/column reference documentation |

## Configuration

Connection via environment variables (see `.env.example`):

| Variable | Default |
|----------|---------|
| `DB_HOST` | `localhost` |
| `DB_PORT` | `5432` |
| `DB_USERNAME` | `postgres` |
| `DB_PASSWORD` | `postgres` |
| `DB_NAME` | `as_ip` |
| `DB_SSLMODE` | `disable` |

## Dependencies

- `internal/config` — `DBConfig` struct and DSN builder
- `internal/repository` — read queries
- `internal/sync` — bulk writes during daily import

## Invariants

- Schema auto-applies only when `country` table does not exist
- `sync_state` row is always `id = 1`
- IP range columns (`start_ip`, `end_ip`) stored as `BIGINT`
- Connection pool: max 25 open, 5 idle
