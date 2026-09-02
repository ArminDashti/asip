# suggestion2 — Retire legacy root deploy scripts

**Component:** `run-on-docker.ps1`, `build-docker-image.ps1`

Preferred path is `.armin/docker-scripts/` (YAML-only). Root scripts still use CLI `--` flags and duplicate deploy logic. Remove or redirect them once callers migrate.

**Effort:** Small
