# Minor Bugs / Risks

## [docker-compose] External network required for bare compose

`docker-compose.yml` marks `asip-net` as external. Bare `docker compose up` fails unless the network exists. `.armin/docker-scripts/run-on-docker-local.ps1` creates it automatically.

## [role table] Unused during import

The `role` table and `asn.network_role_id` column exist in schema but are never populated during sync.
