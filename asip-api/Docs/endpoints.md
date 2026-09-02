# API Endpoints

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| GET | `/api/v1/docs` | Document-like catalog of all API endpoints | No |
| GET | `/api/v1/health` | Health check | No |
| GET | `/api/v1/ip/info` | IP lookup for caller | No |
| GET | `/api/v1/ip/info/:ip` | IP lookup by address | No |
| GET | `/api/v1/ip/whois` | IP whois (RDAP) for caller | No |
| GET | `/api/v1/ip/whois/:ip` | IP whois (RDAP) by address | No |
| GET | `/api/v1/http/headers` | Echo request HTTP headers + client IP | No |
| GET | `/api/v1/dns/lookup/*domain` | DNS lookup (A/AAAA/NS/MX/TXT/CNAME) + ASN/country for resolved IPs | No |
| GET | `/api/v1/asn/list` | List all ASN numbers and handles | No |
| GET | `/api/v1/asn/search/:asn` | ASN details by number or handle | No |
| GET | `/api/v1/asn/to-as/:asn` | Map ASN to AS summary | No |
| GET | `/api/v1/as/search/:as` | AS details by handle or number | No |
| GET | `/api/v1/country/list` | List all countries | No |
| GET | `/api/v1/country/search/:country` | ASN list by ISO country code | No |

## CLI Commands

| Command | Description |
|---------|-------------|
| `asip status` | Show last sync time and request counts |
| `asip sync` | Run a one-off data sync |
