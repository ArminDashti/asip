# Modules

| Module | Role |
|--------|------|
| `App` | Pathname routing (`/`, `/ip`, `/nslookup`) |
| `HomePage` | Auto-load caller IP info |
| `IpPage` | Manual IP lookup form + results |
| `NslookupPage` | Domain DNS lookup form + summary + addresses table |
| `asIpClient` | HTTP client for as-ip `/ip/info` and `/dns/lookup` endpoints |
| `countryFlag` | Country name → `flag-icons` CSS class |
| Styles | Global layout, fields, lookup form, addresses table |
| Deploy | Docker + nginx SPA |
