# Deployment Scripts

| File | Description |
|------|-------------|
| `docker-compose.yml` | Runs PostgreSQL 16 + API service |
| `Dockerfile` | Multi-stage Go build (Alpine) |
| `.armin/docker-scripts/run-on-docker-local.ps1` | Local Docker daemon deploy (YAML-only) |
| `.armin/docker-scripts/run-on-docker-local.yaml` | Local stack settings |
| `.armin/docker-scripts/run-on-docker-server.ps1` | Remote SSH deploy (YAML-only) |
| `.armin/docker-scripts/run-on-docker-server.yaml` | Remote stack + SSH settings |
| `build-docker-image.ps1` | Legacy build/save helper |
| `run-on-docker.ps1` | Legacy CLI-flag deploy script |

## Preferred deploy

```powershell
.\.armin\docker-scripts\run-on-docker-local.ps1
```

Remote: fill `ssh` and `volume_dir` in `run-on-docker-server.yaml`, then:

```powershell
.\.armin\docker-scripts\run-on-docker-server.ps1
```

Compose override env vars from YAML: `IMAGE_TAG`, `DOCKER_NETWORK`, `INTERNAL_PORT`, `PUBLISH_PORT`. Local `publish_port` is `3000` (app default; freed by stack teardown before bind). Server leaves `publish_port` empty when behind a reverse proxy. Both YAML files set `delete_image: "yes"`.

## Docker Compose services

- **postgres** — PostgreSQL 16 with health check, volume `asip-pgdata`
- **asip-api** — Go API, depends on healthy postgres, volume `asip-data` for git clones

## First-time setup

Local deploy script creates `asip-net` if missing. Manual equivalent:

```bash
docker network create asip-net
docker compose up --build
```
