# Directory Tree

```
as-ip/
├── .armin/
│   └── docker-scripts/
│       ├── run-on-docker-local.ps1    # Local Docker deploy (YAML-only)
│       ├── run-on-docker-local.yaml   # Local stack/image/network settings
│       ├── run-on-docker-server.ps1   # Remote SSH Docker deploy
│       └── run-on-docker-server.yaml  # Remote settings + SSH placeholders
├── cmd/
│   ├── api/main.go              # HTTP API server entrypoint
│   ├── asip/main.go             # CLI (status, sync)
│   └── sync/main.go             # Standalone sync entrypoint
├── db/
│   ├── init.sql                 # PostgreSQL schema DDL
│   └── db.md                    # Table/column reference docs
├── internal/
│   ├── config/config.go         # Environment-based configuration
│   ├── database/
│   │   ├── postgres.go          # PostgreSQL connection and schema init
│   │   ├── bulk.go              # Batched bulk INSERT helper
│   │   └── iprange.go           # IPv4 to integer conversion
│   ├── dto/
│   │   ├── as.go                # AS response DTOs
│   │   ├── asn.go               # ASN response DTOs
│   │   ├── country.go           # Country response DTOs
│   │   ├── docs.go              # Docs catalog response DTOs
│   │   ├── health.go            # Health + error response DTOs
│   │   └── ip.go                # IP lookup response DTOs
│   ├── handler/
│   │   ├── docs_handler.go      # GET /api/v1/docs catalog
│   │   ├── error_response.go    # Shared HTTP error helper
│   │   ├── health_handler.go    # GET /api/v1/health
│   │   └── lookup_handler.go    # IP/ASN/AS/country handlers
│   ├── mapper/                  # Domain to DTO mapping
│   ├── model/                   # Domain models
│   ├── repository/              # SQL read/write queries
│   ├── router/router.go         # Gin route registration
│   ├── service/                 # Business logic layer
│   └── sync/                    # Git pull + DB import + scheduler
├── Docs/
│   ├── description.md           # Project overview
│   ├── endpoints.md             # API + CLI endpoint list
│   ├── dir-tree.md              # Annotated directory tree
│   └── modules/
│       ├── database.md          # Database module docs
│       ├── deployment-scripts.md# Docker deploy scripts
│       └── http-handlers.md     # HTTP handler module docs
├── docker-compose.yml           # API + PostgreSQL services
├── Dockerfile                   # Multi-stage Go build
├── build-docker-image.ps1       # Legacy image build helper
├── run-on-docker.ps1            # Legacy CLI-flag deploy
├── go.mod                       # Go module definition
├── .env.example                 # Environment variable template
└── README.md                    # Project overview and setup
```
