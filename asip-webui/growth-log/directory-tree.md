# Directory tree

| Path | Description |
|------|-------------|
| `src/App.tsx` | Pathname switch for home, IP lookup, DNS lookup |
| `src/main.tsx` | React entry |
| `src/pages/HomePage.tsx` | Caller IP summary |
| `src/pages/IpPage.tsx` | Manual IP lookup page |
| `src/pages/NslookupPage.tsx` | Domain DNS lookup page |
| `src/api/asIpClient.ts` | as-ip HTTP client |
| `src/lib/countryFlag.ts` | Country → flag-icons CSS class |
| `src/styles/global.css` | Global styles |
| `nginx.conf` | SPA static hosting |
| `Dockerfile` / `docker-compose.yml` | Container build/run |
| `.armin/docker-scripts/` | Deploy scripts (local/server) |
| `docs/superpowers/specs/` | Design specs |
| `docs/superpowers/plans/` | Implementation plans |
| `growth-log/` | Project growth documentation |
| `README.md` | Setup and API notes |
