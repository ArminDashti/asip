# DNS Lookup Endpoint Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `GET /api/v1/dns/lookup/*domain` that returns live DNS records (A/AAAA/NS/MX/TXT/CNAME) plus ASN/AS/country enrichment for resolved addresses.

**Architecture:** Gin catch-all route → `LookupHandler.LookupDns` → `LookupService.LookupDns` normalizes the host, resolves DNS via Go `net` package, then best-effort enriches each A/AAAA IP using existing `FindByIP` + `FindCountryByIP` (IPv4 only today). Top-level `asn`/`as`/`country` come from the primary enriched address.

**Tech Stack:** Go 1.25, Gin, existing `LookupService` / `AsRepository` / dto JSON patterns. No new dependencies.

**Spec:** `docs/superpowers/specs/2026-07-28-dns-lookup-endpoint-design.md`

## Global Constraints

- Module path: `github.com/ArminDashti/as-ip/server`
- Follow existing handler → service → dto → router → docs_handler patterns
- Reuse `service.ErrBadRequest` / `service.ErrNotFound` and `respondWithError`
- Do not shell out to `nslookup`; use Go `net` DNS APIs only
- Catch-all route `*domain` so `http://` and `https://` inputs work
- IPv6 addresses appear in `aaaa` / `addresses`, but ASN/country enrichment uses current IPv4-only repository helpers (empty `asn`/`as`/`country` for IPv6 until a separate IPv6 find exists — do not expand scope here)
- Prefer commits only when the user asks; plan steps that say “Commit” are optional unless explicitly requested

---

## File map

| File | Responsibility |
|------|----------------|
| `internal/dto/dns.go` | Response types |
| `internal/service/normalize.go` | Add `normalizeDomain` |
| `internal/service/normalize_test.go` | Domain normalize unit tests |
| `internal/service/lookup_dns.go` | `LookupDns` orchestration |
| `internal/handler/lookup_handler.go` | HTTP handler |
| `internal/router/router.go` | Route registration |
| `internal/handler/docs_handler.go` | Docs catalog entry |
| `README.md`, `Docs/endpoints.md`, `Endpoints.md`, `growth-log/features.md` | User-facing docs |

---

### Task 1: DTOs + domain normalization (TDD)

**Files:**
- Create: `internal/dto/dns.go`
- Modify: `internal/service/normalize.go`
- Modify: `internal/service/normalize_test.go`

**Interfaces:**
- Produces: `dto.DnsLookupResponse`, `dto.DnsMxRecord`, `dto.DnsAddressInfo`
- Produces: `func normalizeDomain(value string) (string, error)`

- [ ] **Step 1: Add DTO types**

Create `internal/dto/dns.go`:

```go
package dto

type DnsMxRecord struct {
	Host string `json:"host"`
	Pref uint16 `json:"pref"`
}

type DnsAddressInfo struct {
	Ip      string `json:"ip"`
	Asn     int    `json:"asn"`
	As      string `json:"as"`
	Country string `json:"country"`
}

type DnsLookupResponse struct {
	Domain    string           `json:"domain"`
	A         []string         `json:"a"`
	AAAA      []string         `json:"aaaa"`
	NS        []string         `json:"ns"`
	MX        []DnsMxRecord    `json:"mx"`
	TXT       []string         `json:"txt"`
	CNAME     string           `json:"cname"`
	Asn       int              `json:"asn"`
	As        string           `json:"as"`
	Country   string           `json:"country"`
	Addresses []DnsAddressInfo `json:"addresses"`
}
```

- [ ] **Step 2: Write failing tests for `normalizeDomain`**

Append to `internal/service/normalize_test.go`:

```go
func TestNormalizeDomain(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "bare host", input: "example.com", want: "example.com"},
		{name: "http url", input: "http://example.com", want: "example.com"},
		{name: "https url", input: "https://example.com", want: "example.com"},
		{name: "https with path", input: "https://example.com/path?x=1", want: "example.com"},
		{name: "typo host", input: "exampl.com", want: "exampl.com"},
		{name: "trailing slash url", input: "https://example.com/", want: "example.com"},
		{name: "uppercase host", input: "Example.COM", want: "example.com"},
		{name: "fqdn trailing dot", input: "example.com.", want: "example.com"},
		{name: "leading slash from gin", input: "/https://example.com", want: "example.com"},
		{name: "empty", input: "  ", wantErr: true},
		{name: "whitespace only after slash", input: "/", wantErr: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := normalizeDomain(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %q", got)
				}
				if !errors.Is(err, ErrBadRequest) {
					t.Fatalf("expected ErrBadRequest, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}
```

- [ ] **Step 3: Run test to verify it fails**

Run: `go test ./internal/service -run TestNormalizeDomain -v`

Expected: FAIL — `normalizeDomain` undefined

- [ ] **Step 4: Implement `normalizeDomain`**

Add to `internal/service/normalize.go` (keep existing helpers). Imports needed: `fmt`, `net/url`, `strings`.

```go
func normalizeDomain(value string) (string, error) {
	normalized := strings.TrimSpace(value)
	normalized = strings.TrimPrefix(normalized, "/")
	normalized = strings.TrimSpace(normalized)
	if normalized == "" {
		return "", fmt.Errorf("domain is required: %w", ErrBadRequest)
	}

	candidate := normalized
	if strings.Contains(candidate, "://") ||
		strings.HasPrefix(strings.ToLower(candidate), "http:") ||
		strings.HasPrefix(strings.ToLower(candidate), "https:") {
		if !strings.Contains(candidate, "://") {
			candidate = strings.Replace(candidate, "http:", "http://", 1)
			candidate = strings.Replace(candidate, "https:", "https://", 1)
		}
		parsed, err := url.Parse(candidate)
		if err != nil || parsed.Hostname() == "" {
			return "", fmt.Errorf("invalid domain: %w", ErrBadRequest)
		}
		candidate = parsed.Hostname()
	} else if strings.Contains(candidate, "/") {
		parsed, err := url.Parse("http://" + candidate)
		if err != nil || parsed.Hostname() == "" {
			return "", fmt.Errorf("invalid domain: %w", ErrBadRequest)
		}
		candidate = parsed.Hostname()
	}

	candidate = strings.TrimSuffix(strings.ToLower(strings.TrimSpace(candidate)), ".")
	if candidate == "" || strings.ContainsAny(candidate, " \t") {
		return "", fmt.Errorf("invalid domain: %w", ErrBadRequest)
	}
	return candidate, nil
}
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `go test ./internal/service -run TestNormalizeDomain -v`

Expected: PASS

- [ ] **Step 6: Commit (optional — only if user requested commits)**

```bash
git add internal/dto/dns.go internal/service/normalize.go internal/service/normalize_test.go
git commit -m "feat: add DNS lookup DTOs and domain normalization"
```

---

### Task 2: `LookupDns` service

**Files:**
- Create: `internal/service/lookup_dns.go`

**Interfaces:**
- Consumes: `normalizeDomain`, `*LookupService.repository`, `dto.DnsLookupResponse`
- Produces: `func (s *LookupService) LookupDns(ctx context.Context, domain string) (dto.DnsLookupResponse, error)`

- [ ] **Step 1: Implement DNS lookup + IPv4 enrichment**

Create `internal/service/lookup_dns.go`:

```go
package service

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/ArminDashti/as-ip/server/internal/dto"
	"github.com/ArminDashti/as-ip/server/internal/repository"
)

func (s *LookupService) LookupDns(ctx context.Context, domain string) (dto.DnsLookupResponse, error) {
	normalizedDomain, err := normalizeDomain(domain)
	if err != nil {
		return dto.DnsLookupResponse{}, err
	}

	lookupCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	response := dto.DnsLookupResponse{
		Domain:    normalizedDomain,
		A:         []string{},
		AAAA:      []string{},
		NS:        []string{},
		MX:        []dto.DnsMxRecord{},
		TXT:       []string{},
		CNAME:     "",
		Addresses: []dto.DnsAddressInfo{},
	}

	if ips, lookupErr := net.DefaultResolver.LookupIP(lookupCtx, "ip", normalizedDomain); lookupErr == nil {
		for _, ip := range ips {
			if v4 := ip.To4(); v4 != nil {
				response.A = append(response.A, v4.String())
			} else {
				response.AAAA = append(response.AAAA, ip.String())
			}
		}
	} else if !isDNSNotFound(lookupErr) {
		return dto.DnsLookupResponse{}, lookupErr
	}

	if nsRecords, lookupErr := net.DefaultResolver.LookupNS(lookupCtx, normalizedDomain); lookupErr == nil {
		for _, ns := range nsRecords {
			response.NS = append(response.NS, ns.Host)
		}
	} else if !isDNSNotFound(lookupErr) {
		return dto.DnsLookupResponse{}, lookupErr
	}

	if mxRecords, lookupErr := net.DefaultResolver.LookupMX(lookupCtx, normalizedDomain); lookupErr == nil {
		for _, mx := range mxRecords {
			response.MX = append(response.MX, dto.DnsMxRecord{Host: mx.Host, Pref: mx.Pref})
		}
	} else if !isDNSNotFound(lookupErr) {
		return dto.DnsLookupResponse{}, lookupErr
	}

	if txtRecords, lookupErr := net.DefaultResolver.LookupTXT(lookupCtx, normalizedDomain); lookupErr == nil {
		response.TXT = append(response.TXT, txtRecords...)
	} else if !isDNSNotFound(lookupErr) {
		return dto.DnsLookupResponse{}, lookupErr
	}

	if cname, lookupErr := net.DefaultResolver.LookupCNAME(lookupCtx, normalizedDomain); lookupErr == nil {
		if cname != normalizedDomain+"." && cname != normalizedDomain {
			response.CNAME = cname
		}
	} else if !isDNSNotFound(lookupErr) {
		return dto.DnsLookupResponse{}, lookupErr
	}

	if len(response.A) == 0 && len(response.AAAA) == 0 &&
		len(response.NS) == 0 && len(response.MX) == 0 &&
		len(response.TXT) == 0 && response.CNAME == "" {
		return dto.DnsLookupResponse{}, fmt.Errorf("no DNS records found for %s: %w", normalizedDomain, ErrNotFound)
	}

	orderedIPs := append(append([]string{}, response.A...), response.AAAA...)
	for _, ip := range orderedIPs {
		response.Addresses = append(response.Addresses, s.enrichDnsAddress(lookupCtx, ip))
	}

	for _, address := range response.Addresses {
		if address.Asn != 0 || address.As != "" || address.Country != "" {
			response.Asn = address.Asn
			response.As = address.As
			response.Country = address.Country
			break
		}
	}
	if response.Asn == 0 && response.As == "" && response.Country == "" && len(response.Addresses) > 0 {
		primary := response.Addresses[0]
		response.Asn = primary.Asn
		response.As = primary.As
		response.Country = primary.Country
	}

	return response, nil
}

func (s *LookupService) enrichDnsAddress(ctx context.Context, ip string) dto.DnsAddressInfo {
	info := dto.DnsAddressInfo{Ip: ip}
	if net.ParseIP(ip).To4() == nil {
		return info
	}

	record, err := s.repository.FindByIP(ctx, ip)
	if err != nil {
		return info
	}

	countryName := ""
	if _, geoCountry, geoErr := s.repository.FindCountryByIP(ctx, ip); geoErr == nil {
		countryName = geoCountry
	} else if errors.Is(geoErr, repository.ErrNotFound) && record.CountryName != nil {
		countryName = *record.CountryName
	} else if geoErr != nil && !errors.Is(geoErr, repository.ErrNotFound) {
		return info
	} else if record.CountryName != nil {
		countryName = *record.CountryName
	}

	info.Asn = record.AsnNumber
	info.As = record.Name
	info.Country = countryName
	return info
}

func isDNSNotFound(err error) bool {
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return dnsErr.IsNotFound
	}
	return false
}
```

- [ ] **Step 2: Compile-check the package**

Run: `go test ./internal/service -count=0`

Expected: OK (compiles; existing tests still pass)

- [ ] **Step 3: Commit (optional)**

```bash
git add internal/service/lookup_dns.go
git commit -m "feat: add LookupDns service with DNS resolve and ASN enrichment"
```

---

### Task 3: Handler, router, docs catalog

**Files:**
- Modify: `internal/handler/lookup_handler.go`
- Modify: `internal/router/router.go`
- Modify: `internal/handler/docs_handler.go`

**Interfaces:**
- Consumes: `LookupService.LookupDns`
- Produces: `func (h *LookupHandler) LookupDns(c *gin.Context)`
- Route: `GET /api/v1/dns/lookup/*domain`

- [ ] **Step 1: Add handler method**

Append to `internal/handler/lookup_handler.go`:

```go
func (h *LookupHandler) LookupDns(c *gin.Context) {
	domain := strings.TrimPrefix(c.Param("domain"), "/")
	response, err := h.service.LookupDns(c.Request.Context(), domain)
	if err != nil {
		respondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, response)
}
```

- [ ] **Step 2: Register route**

In `internal/router/router.go`, inside the `api` group (alongside other lookups):

```go
api.GET("/dns/lookup/*domain", lookupHandler.LookupDns)
```

- [ ] **Step 3: Add docs catalog entry**

In `internal/handler/docs_handler.go`, append an `EndpointDocument` (before the closing of `Endpoints`):

```go
{
	Method:      "GET",
	Path:        "/api/v1/dns/lookup/*domain",
	Summary:     "DNS lookup for a domain",
	Description: "Resolves DNS records (A, AAAA, NS, MX, TXT, CNAME) for a website or domain and enriches resolved IPs with ASN and country when available.",
	Auth:        false,
	Parameters: []dto.EndpointParameter{
		{
			Name:        "domain",
			In:          "path",
			Required:    true,
			Description: "Hostname or URL (e.g. example.com, https://example.com)",
		},
	},
},
```

- [ ] **Step 4: Build the API binary**

Run: `go build -o bin/as-ip-api ./cmd/api`

Expected: exit 0

- [ ] **Step 5: Commit (optional)**

```bash
git add internal/handler/lookup_handler.go internal/router/router.go internal/handler/docs_handler.go
git commit -m "feat: expose DNS lookup endpoint and docs catalog entry"
```

---

### Task 4: Project documentation + growth log

**Files:**
- Modify: `README.md` (API endpoints table)
- Modify: `Docs/endpoints.md`
- Modify: `Endpoints.md` (add a DNS Lookup section with example JSON matching the confirmed response)
- Modify: `growth-log/features.md`

- [ ] **Step 1: Update endpoint tables**

Add row:

`GET | /api/v1/dns/lookup/*domain | DNS lookup (A/AAAA/NS/MX/TXT/CNAME) + ASN/country for resolved IPs`

- [ ] **Step 2: Document accepted inputs and example response**

In `Endpoints.md`, document that these work:

- `example.com`
- `http://example.com`
- `https://example.com`
- `exampl.com`

Include the confirmed example JSON from the spec (domain, a, aaaa, ns, mx, txt, cname, asn, as, country, addresses).

- [ ] **Step 3: Update `growth-log/features.md`**

Add: DNS lookup endpoint with live resolver + ASN/country enrichment.

- [ ] **Step 4: Commit (optional)**

```bash
git add README.md Docs/endpoints.md Endpoints.md growth-log/features.md
git commit -m "docs: document DNS lookup endpoint"
```

---

### Task 5: Manual verification

**Files:** none (runtime check)

- [ ] **Step 1: Ensure API can run** (local or Docker with DB synced enough for ASN enrichment on a known IPv4)

- [ ] **Step 2: Hit the endpoint**

```bash
curl -s "http://localhost:3000/api/v1/dns/lookup/example.com"
curl -s "http://localhost:3000/api/v1/dns/lookup/https://example.com"
curl -s "http://localhost:3000/api/v1/dns/lookup/http://example.com"
```

Expected: `200` JSON with `domain` = `example.com`, non-empty `a` and/or `ns`/`txt` as DNS provides, `addresses` present; ASN/country populated when the resolved IPv4 exists in DB.

- [ ] **Step 3: Invalid domain**

```bash
curl -s "http://localhost:3000/api/v1/dns/lookup/%20"
```

Expected: `400` with `error` = `bad_request`

- [ ] **Step 4: Confirm docs catalog**

```bash
curl -s "http://localhost:3000/api/v1/docs"
```

Expected: includes `/api/v1/dns/lookup/*domain`

---

## Spec coverage checklist

| Spec requirement | Task |
|------------------|------|
| Catch-all `*domain` + URL/host normalize | 1, 3 |
| Live A/AAAA/NS/MX/TXT/CNAME | 2 |
| Top-level + per-IP asn/as/country | 2 |
| 400 / 404 / 500 via existing errors | 2, 3 |
| Docs catalog + README/Endpoints | 3, 4 |
| Confirmed response JSON shape | 1 (DTO), 4 (docs) |

## Self-review notes

- No placeholders left in tasks
- IPv6 enrichment limitation called out in Global Constraints (matches current `FindByIP`)
- Type names consistent: `DnsLookupResponse`, `LookupDns`, `normalizeDomain`
