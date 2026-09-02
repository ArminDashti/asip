# NSLookup page design

Date: 2026-07-28  
Status: approved (revised 2026-07-29 — summary Flag/IP/ASN/AS/Country + heading-less Addresses + NS grids)

## Goal

Ship a working `/nslookup` experience:

1. **as-ip** implements `GET /api/v1/dns/lookup/*domain` (DNS resolve + ASN/country enrichment).
2. **asip-webui** shows a domain form, a short summary (**Flag**, **IP**, **ASN**, **AS**, **Country**), then **Addresses** and **NS** grids (no section headings).

## Approach

- Implement the as-ip endpoint per `as-ip/docs/superpowers/specs/2026-07-28-dns-lookup-endpoint-design.md` (Go `net` resolver + existing IP→ASN/geo repository path). Rejected: shelling out to `nslookup`; browser-only resolve + N× `/ip/info` (CORS / fragile).
- Summary `<dl>` matches `/` and `/ip`: Flag (from top-level `country`), **IP** (joined `a[]`, or `—` if empty), **ASN**, **AS**, and **Country**.
- Keep **Addresses** table and **NS** table below the summary without `<h2>` headings.
- Addresses column mapping (UI label ← API field):

| UI column | API field | Display |
|-----------|-----------|---------|
| IP | `ip` | literal address |
| ASN | `asn` | number only (e.g. `15133`) |
| AS | `as` | name only (e.g. `EDGECAST`) |
| Country | `country` | flag (`flag-icons` via existing helper) + country name |

Empty name / zero ASN / empty country → show `—`. Unknown country name → text without flag.

## Architecture

```
Browser (/nslookup)
  → App (pathname switch)
      → NslookupPage
          → asIpClient.fetchDnsLookup(domain)
               GET {VITE_AS_IP_BASE_URL}/dns/lookup/{encodedDomain}
          → summary: Flag + IP (a[]) + ASN + AS + Country
          → addresses[] table + ns[] table (no headings)
               → countryFlag helper per address row

as-ip
  GET /api/v1/dns/lookup/*domain
    → LookupHandler.LookupDns
    → LookupService.LookupDns (normalize → DNS → enrich addresses)
```

## API contract (consumed by UI)

Full response shape is defined in the as-ip DNS lookup design. The UI uses:

- Summary: `a[]`, `asn`, `as`, `country`
- Addresses grid: `addresses[]`
- NS grid: `ns[]`

```json
{
  "a": ["93.184.216.34"],
  "ns": ["a.iana-servers.net", "b.iana-servers.net"],
  "asn": 15133,
  "as": "EDGECAST",
  "country": "United States",
  "addresses": [
    {
      "ip": "93.184.216.34",
      "asn": 15133,
      "as": "EDGECAST",
      "country": "United States"
    }
  ]
}
```

Client keeps typing the full response for forward compatibility.

## Components

| Unit | Responsibility |
|------|----------------|
| as-ip `LookupDns` (dto/service/handler/router/docs) | Endpoint + enrichment |
| `asIpClient.fetchDnsLookup` | Path matches catch-all URL encoding |
| `NslookupPage` | Domain form; loading/error/ready; summary + Addresses + NS |
| `buildCountryFlagClassName` | Country name → `fi fi-xx` |
| `App` | `/nslookup` route (already present) |

## Data flow

1. User submits non-empty trimmed domain (else inline validation, no request).
2. `GET …/dns/lookup/{encodedDomain}` (support bare host and URL-like inputs via API normalize).
3. Success → summary (Flag / IP / ASN / AS / Country) + Addresses rows + NS rows.
4. Empty `addresses` / `ns` → short empty message for that grid.
5. Failure → show API `message` when present.

## Error handling

| Case | Behavior |
|------|----------|
| Empty submit | Inline error, no API call |
| `400` invalid domain | Show API message |
| `404` no DNS records | Show API message |
| Network / other | Generic or API message (with client retries for transient fetch failures) |
| Missing enrichment | `—` in ASN/AS/Country cells |

## Out of scope

- React Router / nav chrome
- Showing CNAME/MX/TXT in the summary
- Custom DNS server parameter
- Lookup history / batch domains
- Changing `/` or `/ip` behavior beyond shared client retries

## Success criteria

1. as-ip serves `GET /api/v1/dns/lookup/example.com` with enriched `addresses`.
2. `/nslookup` shows Flag + IP/ASN/AS/Country, then Addresses and NS grids without section headings.
3. `/` and `/ip` unchanged in layout; empty submit and API errors handled cleanly.
4. Endpoint documented in as-ip `/api/v1/docs` and endpoint docs.

## Manual test

1. Deploy/run as-ip with DB sync; curl DNS lookup for `example.com` → `addresses` with ASN/country when known.
2. Open asip-webui `/nslookup`, look up same domain → summary + two grids.
3. Empty submit → validation; nonsense domain → error; `/` and `/ip` still work.
