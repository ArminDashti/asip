# DNS Lookup Endpoint Design

**Date:** 2026-07-28  
**Status:** Approved (response shape confirmed 2026-07-28)  
**Service:** as-ip (Go + Gin API)

## Goal

Add a public GET endpoint that accepts a website/domain and returns DNS records as structured JSON (nslookup-style data without shelling out to the CLI), plus **ASN** and **country** for resolved addresses using the existing IP→ASN/geo data (same source as `GET /api/v1/ip/info/:ip`).

## Approach

1. Resolve DNS with Go’s standard library (`net.DefaultResolver` / `LookupIP`, `LookupNS`, `LookupMX`, `LookupTXT`, `LookupCNAME`).
2. Enrich each resolved A/AAAA address via the existing repository path used by `GetIPInfo` (`FindByIP` + `FindCountryByIP`).
3. Expose top-level `asn`, `as`, and `country` from the **primary** address (first A with a mapping, else first AAAA with a mapping, else first A, else first AAAA).

No new dependencies. Follow the existing handler → service → dto pattern.

Rejected alternatives:

- Shelling out to `nslookup` (platform-fragile, harder to parse, weaker security)
- Third-party DNS libraries (unnecessary for this scope)

## Endpoint

| Method | Path | Auth |
|--------|------|------|
| GET | `/api/v1/dns/lookup/*domain` | No |

Examples:

- `GET /api/v1/dns/lookup/example.com`
- `GET /api/v1/dns/lookup/http://example.com`
- `GET /api/v1/dns/lookup/https://example.com`
- `GET /api/v1/dns/lookup/exampl.com`

Use Gin’s **catch-all** `*domain` (not `:domain`) so scheme-prefixed URLs with `/` still reach the handler. Gin typically returns the param with a leading `/` (e.g. `/https://example.com`); the handler strips that before calling the service.

Practical accepted inputs (string received by the normalizer):

| Input | Normalized domain |
|-------|-------------------|
| `example.com` | `example.com` |
| `http://example.com` | `example.com` |
| `https://example.com` | `example.com` |
| `exampl.com` | `exampl.com` |
| `https://example.com/` | `example.com` |
| `https://example.com/path?x=1` | `example.com` |

The service layer always strips scheme, credentials, port, path, query, and fragment when present.

### Query / body

None. Domain is path-only via catch-all (same family as other lookup paths, but wildcard so URL inputs work).

## Input normalization

Service-layer function (e.g. `normalizeDomain`):

1. Trim whitespace
2. Reject empty input → `400 bad_request`
3. If the string looks like a URL (contains `://` or starts with `http:` / `https:`), parse with `net/url` and take `Hostname()`
4. Otherwise treat as host: strip optional leading `http://` / `https://` if present without full parse edge cases
5. Strip trailing `.` (FQDN dot) and lowercase the hostname for consistent output
6. Reject if hostname is empty, contains spaces, or is clearly not a DNS name (invalid characters)
7. Do **not** execute shell commands; never pass the input to an OS process

Notes:

- Typo domains such as `exampl.com` are valid input; lookup proceeds and returns whatever DNS exists (or not-found if none).
- Ports in host (`example.com:443`) are stripped via `Hostname()` / host parsing.

## Response

### Success — `200 OK`

```json
{
  "domain": "example.com",
  "a": ["93.184.216.34"],
  "aaaa": ["2606:2800:220:1:248:1893:25c8:1946"],
  "ns": ["a.iana-servers.net.", "b.iana-servers.net."],
  "mx": [{"host": "mail.example.com.", "pref": 10}],
  "txt": ["v=spf1 -all"],
  "cname": "",
  "asn": 15133,
  "as": "EDGECAST",
  "country": "United States",
  "addresses": [
    {
      "ip": "93.184.216.34",
      "asn": 15133,
      "as": "EDGECAST",
      "country": "United States"
    },
    {
      "ip": "2606:2800:220:1:248:1893:25c8:1946",
      "asn": 15133,
      "as": "EDGECAST",
      "country": "United States"
    }
  ]
}
```

Rules:

- `domain` is the normalized hostname used for lookup
- Missing record types use empty arrays (`[]`) or empty string for `cname` — not omitted, not errors
- `a` / `aaaa` are split from `LookupIP` by address family
- `addresses` lists every A/AAAA IP with best-effort `asn` / `as` / `country` (same field meanings as `/ip/info`)
- Top-level `asn`, `as`, `country` mirror the **primary** address (see Approach)
- If an IP has no ASN/geo mapping in the DB, that address still appears with `asn: 0`, `as: ""`, `country: ""` — DNS success is not turned into `404` solely because enrichment is missing
- If there are no A/AAAA addresses, `addresses` is `[]` and top-level `asn`/`as`/`country` are zero/empty; other DNS records (NS/MX/TXT/CNAME) may still yield `200`
- Partial DNS success is OK: if A exists but MX is empty, still return found records
- `404` only when **no** useful DNS records are returned across all looked-up types (no A/AAAA/NS/MX/TXT and empty CNAME)

### Errors

| Status | Condition |
|--------|-----------|
| `400` | Invalid / empty domain after normalization (`service.ErrBadRequest`) |
| `404` | Domain syntactically OK but no DNS records found (`service.ErrNotFound`) |
| `500` | Unexpected resolver / internal failure |

Error body matches existing `dto.ErrorResponse` (`error`, `message`).

## Components

| Layer | Change |
|-------|--------|
| `internal/dto` | Add `DnsLookupResponse`, `DnsMxRecord`, `DnsAddressInfo` |
| `internal/service` | Add domain normalize + `LookupDns` on `LookupService` (DNS resolve + reuse IP ASN/country lookup helpers/repository) |
| `internal/handler` | Add `LookupDns` on `LookupHandler` |
| `internal/router` | Register `GET /dns/lookup/*domain` under `/api/v1` |
| `internal/handler/docs_handler.go` | Catalog entry |
| Docs / README | Document endpoint, accepted inputs, ASN/country fields |

No schema/sync/config changes — enrichment uses existing ASN/prefix/geo tables.

## Data flow

```
Client GET /api/v1/dns/lookup/*domain
  → LookupHandler.LookupDns (strip leading "/" from catch-all param)
  → LookupService.LookupDns(ctx, domain)
       → normalizeDomain(domain)
       → sequential net lookups (A/AAAA via LookupIP, NS, MX, TXT, CNAME)
       → for each A/AAAA: repository FindByIP + FindCountryByIP (best-effort)
       → set top-level asn/as/country from primary enriched address
       → assemble DnsLookupResponse
  → JSON 200 | respondWithError
```

Request logging middleware already covers non-health/docs paths; no special-case needed.

## Error handling & timeouts

- Use `c.Request.Context()` for all lookups so client disconnect cancels work
- Prefer a bounded lookup timeout if the request context has none (e.g. derive with a short timeout such as 5s) to avoid hanging workers
- Map DNS “not found” / empty aggregate results to `ErrNotFound`
- Map invalid input to `ErrBadRequest`
- Do not leak internal resolver details in `500` messages (existing handler pattern)

## Testing

- Unit tests for `normalizeDomain`: bare host, `http://`, `https://`, trailing slash/path/query when present in string, typo host, empty, invalid
- Optional unit/integration test for service with a fake resolver interface if introduced; otherwise keep resolver behind a small interface for testability, or test normalize only in v1

## Out of scope

- Custom nameserver query parameter
- SOA / SRV / PTR / reverse lookup
- Rate limiting / abuse controls beyond existing request logging
- Changing CORS method set (remains GET + OPTIONS)

## Success criteria

1. `GET /api/v1/dns/lookup/example.com` returns structured DNS JSON including top-level `asn`, `as`, `country` and per-IP `addresses`
2. `http://example.com`, `https://example.com`, and bare hosts like `exampl.com` work as path inputs (catch-all route + normalize)
3. ASN/country come from the same data path as `/ip/info` for resolved IPs
4. Endpoint appears in `/api/v1/docs` and project endpoint docs
5. Invalid domain → 400; unknown domain with no DNS records → 404
