# Suggestion: Keep docs catalog and routes in sync

The `GET /api/v1/docs` catalog in `internal/handler/docs_handler.go` is a static list. When routes change in `internal/router/router.go`, the catalog can drift.

**Why it matters:** Clients relying on `/docs` would see stale paths or miss new endpoints.

**Rough effort:** Low — either add a unit test that asserts registered Gin routes match the catalog paths, or generate the catalog from a shared route definition table.
