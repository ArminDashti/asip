# NSLookup Page Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `/nslookup` page that looks up a domain via as-ip DNS API and shows summary fields plus an addresses grid (IP, AS, ASN, Country+Flag).

**Architecture:** Pathname switch in `App`. Extend `asIpClient` with `fetchDnsLookup`. New `NslookupPage` mirrors `IpPage` form/state; summary `<dl>` + HTML `<table>` for `addresses`. Omit `aaaa` / `mx` / `txt` from the UI.

**Tech Stack:** React 19, Vite 7, TypeScript, existing `fetch` client, `flag-icons`.

## Global Constraints

- Keep `/` and `/ip` behavior unchanged
- No react-router dependency
- Match existing CSS / brand
- No nav links in v1
- Unknown paths (except `/`, `/ip`, `/nslookup`) → `location.replace('/')`
- UI must not display `aaaa`, `mx`, or `txt`

---

## File map

| File | Change |
|------|--------|
| `src/api/asIpClient.ts` | Add `DnsLookupResult` / `DnsAddress` types + `fetchDnsLookup` |
| `src/pages/NslookupPage.tsx` | New page |
| `src/App.tsx` | Route `/nslookup` |
| `src/styles/global.css` | Addresses table styles |
| `README.md` | Document `/nslookup` |
| `growth-log/*` | Update after feature |

---

### Task 1: API client

**Files:**
- Modify: `src/api/asIpClient.ts`
- Consumes: existing `resolveApiBaseUrl`, `ApiErrorBody` error pattern
- Produces: `DnsAddress`, `DnsLookupResult`, `fetchDnsLookup(domain: string): Promise<DnsLookupResult>`

- [x] **Step 1: Add types and fetch helper**
- [x] **Step 2: Commit** (batched with feature commit)

---

### Task 2: NslookupPage + styles + route

**Files:**
- Create: `src/pages/NslookupPage.tsx`
- Modify: `src/App.tsx`, `src/styles/global.css`
- Consumes: `fetchDnsLookup`, `DnsLookupResult`, `buildCountryFlagClassName`
- Produces: `/nslookup` page with summary + addresses table

- [x] **Step 1: Create `NslookupPage`**
- [x] **Step 2: Add table CSS**
- [x] **Step 3: Wire `/nslookup` in `App.tsx`**
- [x] **Step 4: Commit** (batched with feature commit)

---

### Task 3: Docs + verify

**Files:**
- Modify: `README.md`, `growth-log/*`

- [x] **Step 1: Document `/nslookup` in README**
- [x] **Step 2: Update growth-log**
- [x] **Step 3: `npm run build` passes**
- [x] **Step 4: Commit**
