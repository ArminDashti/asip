# Architecture (schematic)

```
Browser
  ├── /          → HomePage      → GET /api/v1/ip/info
  ├── /ip        → IpPage        → GET /api/v1/ip/info/{ip}
  └── /nslookup  → NslookupPage  → GET /api/v1/dns/lookup/{domain}
                │
                ├── asIpClient
                └── countryFlag → flag-icons (CSS)
                         │
                    as-ip API (VITE_AS_IP_BASE_URL)
```
