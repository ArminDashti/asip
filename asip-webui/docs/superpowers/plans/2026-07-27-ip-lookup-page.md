# IP Lookup Page Implementation Plan

> **For agentic workers:** Implement task-by-task. Steps use checkbox syntax.

**Goal:** Add `/ip` page with IP input that returns AS, ASN, Country + flag via as-ip API.

**Architecture:** Pathname switch in `App` (no react-router). Extend `asIpClient` with path-based lookup. New `IpPage` reuses home field/flag patterns.

**Tech Stack:** React 19, Vite 7, existing as-ip API (`GET /ip/info/{ip}`).

## Global Constraints

- Keep home `/` behavior unchanged
- No react-router dependency
- Match existing CSS / brand
- No nav links in v1
- Unknown paths → `location.replace('/')`

---

## File map

| File | Change |
|------|--------|
| `src/api/asIpClient.ts` | Add `fetchIpInfoByAddress(ip)` |
| `src/pages/IpPage.tsx` | New lookup page |
| `src/App.tsx` | Pathname routing |
| `src/styles/global.css` | Form/box styles for lookup |
| `nginx.conf` | Confirm SPA fallback for `/ip` |
| `README.md` | Document `/ip` and new API call |
| `growth-log/*` | Update after feature |

---

### Task 1: API client

- [x] Add `fetchIpInfoByAddress(ip: string)` → `GET .../ip/info/${encodeURIComponent(ip)}`
- [x] Reuse existing error parsing from `fetchIpInfo`

### Task 2: IpPage + styles

- [x] Create `IpPage` with input, submit, loading/error/ready states
- [x] Show ASN, AS, Country + flag (and echoed IP)
- [x] Add minimal form CSS matching home

### Task 3: Routing

- [x] `App` switches on pathname `/` vs `/ip`
- [x] Other paths redirect to `/`
- [x] Confirm nginx `try_files` SPA fallback

### Task 4: Docs + verify

- [x] Update README API/routes note
- [x] Update growth-log
- [x] `npm run build` passes
