# HTTP Handler Module

Gin handlers that expose the AS-IP lookup API and meta endpoints.

## Responsibility

- Map HTTP requests to lookup service calls
- Return JSON DTOs for IP, ASN, AS, and country queries
- Expose health and API documentation catalog endpoints
- Translate service errors into consistent HTTP error responses

## Key files

| File | Description |
|------|-------------|
| `internal/handler/docs_handler.go` | `GET /api/v1/docs` endpoint catalog |
| `internal/handler/health_handler.go` | `GET /api/v1/health` liveness response |
| `internal/handler/lookup_handler.go` | IP / ASN / AS / country lookup handlers |
| `internal/handler/error_response.go` | Shared error JSON responses |
| `internal/router/router.go` | Route registration and request logging |
| `internal/dto/docs.go` | Docs catalog response types |

## Dependencies

- `internal/service` — lookup business logic
- `internal/dto` — response shapes
- `internal/repository` — request logging from the router

## Invariants

- `/api/v1/health` and `/api/v1/docs` are excluded from request logging
- Docs catalog is maintained in `docs_handler.go` and should stay aligned with registered routes
