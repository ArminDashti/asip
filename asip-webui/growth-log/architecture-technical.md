# Architecture (technical)

Vite + React 19 SPA. No client router package: `App` reads `window.location.pathname` and renders `HomePage`, `IpPage`, or `NslookupPage`. Unknown paths replace to `/`.

API base comes from `VITE_AS_IP_BASE_URL`. Shared IP/DNS types and response error parsing live in `asIpClient` (`/ip/info`, `/ip/info/{ip}`, `/dns/lookup/{domain}`). Country flags are client-side name→ISO maps rendered with the bundled `flag-icons` CSS classes (`fi fi-xx`).

Production assets are served by nginx with `try_files` SPA fallback so `/ip` and `/nslookup` deep links work.
