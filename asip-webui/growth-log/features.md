# Features

- **Caller IP summary (`/`)** — On load, fetch the visitor’s IP, ASN, AS, and country from as-ip; show country flag when mappable.
- **IP lookup (`/ip`)** — Form to enter an IP; lookup returns ASN, AS, country, and flag via `GET /ip/info/{ip}`.
- **IP whois (`/whois`)** — Form (or My IP) for RDAP whois via `GET /ip/whois` and `GET /ip/whois/{ip}`.
- **DNS lookup (`/nslookup`)** — Form to enter a domain; shows domain, A, NS, CNAME, ASN, AS, country+flag, plus an addresses table (IP, AS, ASN, Country+Flag) via `GET /dns/lookup/{domain}` (omits AAAA/MX/TXT in the UI).
- **DNS leak (`/dns-leak`)** — Browser DoH whoami probes across public resolvers; compare discovered IPs to caller IP; optional server DNS cross-check.
- **WebRTC leak (`/webrtc`)** — Browser ICE candidate gather; compare public candidates to caller IP.
- **HTTP headers (`/headers`)** — Show headers and client IP as seen by the API via `GET /http/headers`.
- **Theme** — Follows system preference by default (`prefers-color-scheme`) with a header cycle for System / Light / Dark; preference persisted in `localStorage`.
- **Country flags** — Map country display names from as-ip to ISO codes and render via the local `flag-icons` CSS package (no external flag CDN).
- **Docker deploy** — Local and server scripts under `.armin/docker-scripts/` with nginx SPA hosting.
