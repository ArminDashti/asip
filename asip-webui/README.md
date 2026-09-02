# asip-webui

Minimal production SPA that shows the caller's IP, ASN, AS, and country from the as-ip API, plus IP and DNS lookup pages.

## Requirements

- Node.js 22+ (local `npm run dev` / `npm run build`)
- Docker (local/server deploy)
- as-ip API with CORS enabled (production: `https://asip-api.xaigrok.ir`)

## Environment

| Variable | Purpose | Default |
|----------|---------|---------|
| `VITE_AS_IP_BASE_URL` | as-ip API base URL | `https://asip-api.xaigrok.ir/api/v1` |

Copy `.env.example` to `.env` for local overrides. Production builds use `.env.production` (and the Dockerfile build-arg).

| Page | Path | API |
|------|------|-----|
| Caller IP | `/` | `GET {VITE_AS_IP_BASE_URL}/ip/info` |
| IP lookup | `/ip` | `GET {VITE_AS_IP_BASE_URL}/ip/info/{ip}` |
| IP whois | `/whois` | `GET {VITE_AS_IP_BASE_URL}/ip/whois` / `.../ip/whois/{ip}` |
| DNS lookup | `/nslookup` | `GET {VITE_AS_IP_BASE_URL}/dns/lookup/{domain}` |
| DNS leak | `/dns-leak` | Browser DoH + `GET .../ip/info` and optional `GET .../dns/lookup/{domain}` |
| WebRTC leak | `/webrtc` | Browser WebRTC + optional `GET .../ip/info` |
| HTTP headers | `/headers` | `GET {VITE_AS_IP_BASE_URL}/http/headers` |

## Local development

```powershell
npm install
npm run dev
```

Open `http://localhost:5173`. The API must allow origin `http://localhost:5173` via `CORS_ALLOWED_ORIGINS` on as-ip.

## as-ip CORS

On the as-ip service, set:

```env
CORS_ALLOWED_ORIGINS=http://localhost:5173,https://asip.xaigrok.ir
```

Uses `github.com/gin-contrib/cors` (GET + OPTIONS). Redeploy as-ip after changing this so `asip-api.xaigrok.ir` sends `Access-Control-Allow-Origin`.

## Docker deploy

Scripts live under `.armin/docker-scripts/` (YAML only; no CLI flags).

### Local

```powershell
# Ensure external networks exist (script creates asip-webui-net; create t3-net if missing)
docker network create t3-net

.\.armin\docker-scripts\run-on-docker-local.ps1
```

UI: `http://localhost:8082`

### Server (`ssh t3 -p 80`)

```powershell
.\.armin\docker-scripts\run-on-docker-server.ps1
```

Config: `ssh: "ssh t3 -p 80"`, `volume_dir: /home/cloud-admin/docker/asip-webui`, empty `publish_port` (HAProxy on `t3-net`).
