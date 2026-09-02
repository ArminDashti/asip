# AS-IP API Endpoints

Base URL: `http://localhost:3000/api/v1`

All endpoints return **JSON**. Error responses use `{ "error": "<code>", "message": "<details>" }`.

Path parameters are **case-insensitive** where noted. Country filters accept **ISO 3166-1 alpha-2 codes only** (e.g. `US`, `us`).

---

## Health

### `GET /health`

Service health check.

**Response 200**

```json
{
  "status": "ok",
  "service": "as-ip"
}
```

---

## IP Lookup

### `GET /ip/info`

Resolve the caller's IPv4 address (from the HTTP request) to the announcing autonomous system.

When no `ip` path parameter is supplied, the server uses the client IP (`ClientIP()` — respects `X-Forwarded-For` / `X-Real-IP` when behind a trusted proxy).

**Response 200**

```json
{
  "ip": "8.8.8.8",
  "asn": 15169,
  "as": "Google LLC",
  "country": "United States"
}
```

**Response 400** — invalid or missing client IP

**Response 404** — no ASN mapping for the address

---

### `GET /ip/info/:ip`

Resolve an IPv4 address to the announcing autonomous system.

| Parameter | Location | Description |
|-----------|----------|-------------|
| `ip` | path | IPv4 address (e.g. `8.8.8.8`) |

**Response 200**

```json
{
  "ip": "8.8.8.8",
  "asn": 15169,
  "as": "Google LLC",
  "country": "United States"
}
```

**Response 400** — invalid or missing IP

**Response 404** — no ASN mapping for the address

---

## IP Whois (RDAP)

### `GET /ip/whois`

Resolve RDAP whois details for the caller's IP address.

**Response 200**

```json
{
  "ip": "8.8.8.8",
  "handle": "NET-8-8-8-0-1",
  "name": "GOGL",
  "type": "ALLOCATION",
  "country": "US",
  "startAddress": "8.8.8.0",
  "endAddress": "8.8.8.255",
  "status": ["active"],
  "entities": ["Google LLC"],
  "events": [{"action": "last changed", "date": "2024-01-01T00:00:00Z"}],
  "remarks": ["Public DNS"],
  "rdapUrl": "https://rdap.arin.net/registry/ip/8.8.8.0/24"
}
```

**Response 400** — invalid or missing client IP

**Response 404** — no RDAP record for the address

---

### `GET /ip/whois/:ip`

Resolve RDAP whois details for an IPv4 or IPv6 address.

| Parameter | Location | Description |
|-----------|----------|-------------|
| `ip` | path | IPv4 or IPv6 address |

**Response 200** — same shape as `/ip/whois`

**Response 400** — invalid IP

**Response 404** — no RDAP record for the address

---

## HTTP Headers

### `GET /http/headers`

Echo the HTTP headers the API received for this request, plus the resolved client IP. `Cookie` and `Authorization` values are redacted. Hop-by-hop headers such as `Connection` are omitted.

**Response 200**

```json
{
  "ip": "203.0.113.10",
  "headers": [
    { "name": "Accept-Language", "value": "en-US" },
    { "name": "User-Agent", "value": "Mozilla/5.0" }
  ]
}
```

---

## DNS Lookup

### `GET /dns/lookup/*domain`

Resolve DNS records for a hostname or URL and enrich resolved IPs with ASN / AS / country when available.

Accepted path values (catch-all so schemes with `/` work):

- `example.com`
- `http://example.com`
- `https://example.com`
- `exampl.com`

| Parameter | Location | Description |
|-----------|----------|-------------|
| `domain` | path | Hostname or URL |

**Response 200**

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
      "asn": 0,
      "as": "",
      "country": ""
    }
  ]
}
```

Note: IPv6 addresses appear in `aaaa` / `addresses`, but ASN/country enrichment currently applies to IPv4 only.

**Response 400** — empty or invalid domain

**Response 404** — no DNS records found

---

## ASN

### `GET /asn/list`

List all ASN numbers with handles.

**Response 200**

```json
{
  "items": [
    { "asn": 15169, "handle": "GOOGLE" }
  ],
  "total": 1
}
```

---

### `GET /asn/search/:asn`

Lookup ASN details by number or registry handle (case-insensitive).

| Parameter | Location | Description |
|-----------|----------|-------------|
| `asn` | path | ASN number (`15169`, `as15169`) or handle (`GOOGLE`, `google`) |

**Response 200** — `AsnResponse`

**Response 404** — ASN not found

---

### `GET /asn/to-as/:asn`

Map an ASN to its parent AS summary record.

| Parameter | Location | Description |
|-----------|----------|-------------|
| `asn` | path | ASN number or handle (case-insensitive) |

**Response 200**

```json
{
  "asn": 15169,
  "as": {
    "id": 1,
    "handle": "GOOGLE",
    "organization": "Google LLC",
    "countryCode": "US",
    "country": "United States"
  }
}
```

**Response 404** — mapping not found

---

## Autonomous System (AS)

### `GET /as/search/:as`

Lookup AS details by handle, organization name fragment, or ASN number (case-insensitive).

| Parameter | Location | Description |
|-----------|----------|-------------|
| `as` | path | Handle (`GOOGLE`), org name, or number (`15169`, `AS15169`) |

**Response 200** — `AsnResponse`

**Response 404** — AS not found

---

## Country

### `GET /country/list`

List all countries (ISO code and name).

**Response 200**

```json
{
  "items": [
    { "code": "US", "name": "United States" }
  ],
  "total": 1
}
```

---

### `GET /country/search/:country`

List ASN records registered in a country.

| Parameter | Location | Description |
|-----------|----------|-------------|
| `country` | path | ISO 3166-1 alpha-2 code only (`US`, `de`) |

**Response 200** — `AsnListResponse`

**Response 400** — invalid country code

---

## Error Responses

| HTTP Status | `error` value | When |
|-------------|---------------|------|
| 400 | `bad_request` | Missing or invalid path parameter |
| 404 | `not_found` | No matching record |
| 500 | `internal_error` | Unexpected server error |

---

## Quick Examples

```bash
curl http://localhost:3000/api/v1/health
curl http://localhost:3000/api/v1/ip/info
curl http://localhost:3000/api/v1/ip/info/8.8.8.8
curl http://localhost:3000/api/v1/asn/list
curl http://localhost:3000/api/v1/asn/search/15169
curl http://localhost:3000/api/v1/asn/to-as/google
curl http://localhost:3000/api/v1/as/search/GOOGLE
curl http://localhost:3000/api/v1/country/list
curl http://localhost:3000/api/v1/country/search/us
```
