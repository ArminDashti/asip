# Suggestions

## suggestion1 — Add database migration framework

**Component:** `internal/database/postgres.go`

Currently schema is applied once by splitting `init.sql` on semicolons. A proper migration tool (e.g. goose, golang-migrate) would handle incremental schema changes safely without re-running the full DDL.

**Effort:** Small–medium
