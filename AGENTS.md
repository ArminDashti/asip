## Learned User Preferences

- After coding tasks that change a Cloudflare Worker, redeploy by default unless the user explicitly says to skip deploy.
- When deploying to Cloudflare, upload only the files needed for that Worker so upload size stays small; do not upload the whole monorepo or unrelated apps.
- For ASIP on Cloudflare, prefer a Cloudflare-hosted database with a fresh start rather than migrating or transferring existing data.

## Learned Workspace Facts

- ASIP is a monorepo with `asip-api` and `asip-webui`.
- Production Workers: WebUI at `https://asip.armindashti.workers.dev`, API at `https://asip-api.armindashti.workers.dev`.
- API and WebUI deploy as separate Workers, so either can be updated independently.
- IP sync uses `SYNC_INTERVAL_DAYS` (default 5): each run purges import tables then re-imports so only the latest IP set is kept.
- Local Docker Compose project name is `asip-local`; the local API container is typically exposed on port 8174.
