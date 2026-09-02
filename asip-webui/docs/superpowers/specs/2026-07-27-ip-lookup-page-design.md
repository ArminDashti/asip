# IP lookup page design

Date: 2026-07-27  
Status: approved for planning

## Goal

Add a page at `/ip` where a user enters an IP address and sees **AS**, **ASN**, and **Country** (with flag). Keep the existing home page (`/`) that shows the caller’s IP unchanged.

## Approach

Lightweight pathname routing in `App` (no `react-router`). Extend the existing as-ip client and reuse the home page’s field/flag presentation patterns.

## Architecture

```
Browser (/ or /ip)
    → App (pathname switch)
        → HomePage        GET /ip/info          (caller)
        → IpPage          GET /ip/info/{ip}     (user-entered)
            → asIpClient
            → countryFlag helper
```

## Components

| Unit | Responsibility |
|------|----------------|
| `App` | Read `window.location.pathname`; render `HomePage` for `/`, `IpPage` for `/ip`; any other path redirects to `/` via `location.replace('/')` |
| `IpPage` | Form (IP input + Lookup), load/error/ready states, result display |
| `asIpClient.fetchIpInfoByAddress(ip)` | `GET {VITE_AS_IP_BASE_URL}/ip/info/{encodedIp}` → `IpInfo` |
| Existing `buildCountryFlagClassName` | Country name → `flag-icons` CSS class (`fi fi-xx`) |

## Data flow

1. User types IP and submits (button or Enter).
2. Client validates that the string is non-empty (optional light IPv4/IPv6 shape check is allowed; API remains source of truth).
3. Call `fetchIpInfoByAddress(trimmedIp)`.
4. On success, show AS, ASN, Country + flag (same `IpInfo` shape as home: `ip`, `asn`, `as`, `country`).
5. On failure, show a short status/error message from the API or a generic fallback.

## UI

- Match home visual language: centered layout, ASIP brand, field labels, mono values, existing CSS variables/animations.
- Primary control: one text input (placeholder e.g. `8.8.8.8`) and a Lookup control in a simple box/form — not a card-heavy layout.
- Result block: flag (if resolvable), then ASN, AS, Country. Showing the resolved IP is optional but recommended for clarity when the API echoes it.
- No nav links between `/` and `/ip` in v1 (URL/bookmark only).

## Error handling

- Empty submit → inline validation message; no API call.
- Non-OK HTTP / network → surface API `message` when present, else a short generic error.
- Unknown country → omit flag; still show country text.

## Out of scope

- React Router / multi-page nav chrome
- History of lookups
- Batch lookup
- Changing home page behavior or API base URL conventions

## Testing (manual)

1. Open `/ip`, look up `8.8.8.8` → Google LLC / ASN 15169 / United States + flag.
2. Invalid or empty input → clear validation/error, no crash.
3. `/` still loads caller IP as today.
4. Direct load of `/ip` works in Vite dev and production static hosting (SPA fallback must serve `index.html` for `/ip` if not already).

## SPA hosting note

If the deploy server does not already fall back unknown paths to `index.html`, `/ip` direct loads will 404. Confirm or add fallback as part of implementation if needed.
