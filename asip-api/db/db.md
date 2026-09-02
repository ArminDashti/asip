# AS-IP Database Schema

PostgreSQL schema for IP/ASN geolocation lookups. Applied automatically on first connection via `db/init.sql`.

## country

| column | datatype | desc |
|--------|----------|------|
| id | SERIAL | Primary key |
| country | TEXT | Full country name |
| country_code | TEXT | ISO 3166-1 alpha-2 code (unique) |

## origin

| column | datatype | desc |
|--------|----------|------|
| id | SERIAL | Primary key |
| origin | TEXT | ASN origin type (unique) |
| description | TEXT | Origin description |

## category

| column | datatype | desc |
|--------|----------|------|
| id | SERIAL | Primary key |
| category | TEXT | ASN category label (unique) |
| description | TEXT | Category description |

## role

| column | datatype | desc |
|--------|----------|------|
| id | SERIAL | Primary key |
| role | TEXT | Network role label (unique) |
| description | TEXT | Role description |

## asn

| column | datatype | desc |
|--------|----------|------|
| id | SERIAL | Primary key |
| asn | INTEGER | Autonomous System Number (unique) |
| as | TEXT | AS handle / organization name |
| country_id | INTEGER | FK → country(id), ON DELETE SET NULL |
| origin_id | INTEGER | FK → origin(id), ON DELETE SET NULL |
| category_id | INTEGER | FK → category(id), ON DELETE SET NULL |
| network_role_id | INTEGER | FK → role(id), ON DELETE SET NULL |
| registered | TEXT | Registration date |
| last_modified | TEXT | Last metadata modification date |
| last_announced | TEXT | Last BGP announcement date |
| prefixes_last_modified | TEXT | Last prefix update date |

**Indexes:** `idx_asn_asn` (asn), `idx_asn_country` (country_id)

## ipv4_stat

| column | datatype | desc |
|--------|----------|------|
| id | SERIAL | Primary key |
| asn_id | INTEGER | FK → asn(id), ON DELETE CASCADE (unique) |
| prefixes | INTEGER | Total IPv4 prefix count |
| prefixes_aggregated | INTEGER | Aggregated IPv4 prefix count |
| largest_prefix | INTEGER | Largest IPv4 prefix length |
| total_addresses | BIGINT | Total IPv4 addresses announced |

## ipv6_stat

| column | datatype | desc |
|--------|----------|------|
| id | SERIAL | Primary key |
| asn_id | INTEGER | FK → asn(id), ON DELETE CASCADE (unique) |
| prefixes | INTEGER | Total IPv6 prefix count |
| prefixes_aggregated | INTEGER | Aggregated IPv6 prefix count |
| largest_prefix | INTEGER | Largest IPv6 prefix length |
| total_addresses | BIGINT | Total IPv6 addresses announced |

## asn_ipv4

| column | datatype | desc |
|--------|----------|------|
| id | SERIAL | Primary key |
| asn_id | INTEGER | FK → asn(id), ON DELETE CASCADE |
| cidr | TEXT | IPv4 CIDR block |
| last_modified | TEXT | Last modification date |
| start_ip | BIGINT | Start of IP range (integer) |
| end_ip | BIGINT | End of IP range (integer) |

**Unique:** (asn_id, cidr) — **Indexes:** `idx_asn_ipv4_asn`, `idx_asn_ipv4_range` (start_ip, end_ip)

## asn_ipv6

| column | datatype | desc |
|--------|----------|------|
| id | SERIAL | Primary key |
| asn_id | INTEGER | FK → asn(id), ON DELETE CASCADE |
| cidr | TEXT | IPv6 CIDR block |
| last_modified | TEXT | Last modification date |

**Unique:** (asn_id, cidr) — **Index:** `idx_asn_ipv6_asn`

## country_ipv4

| column | datatype | desc |
|--------|----------|------|
| id | SERIAL | Primary key |
| country_id | INTEGER | FK → country(id), ON DELETE CASCADE |
| cidr | TEXT | IPv4 CIDR block |
| last_modified | TEXT | Last modification date |
| start_ip | BIGINT | Start of IP range (integer) |
| end_ip | BIGINT | End of IP range (integer) |

**Unique:** (country_id, cidr) — **Indexes:** `idx_country_ipv4_country`, `idx_country_ipv4_range` (start_ip, end_ip)

## country_ipv6

| column | datatype | desc |
|--------|----------|------|
| id | SERIAL | Primary key |
| country_id | INTEGER | FK → country(id), ON DELETE CASCADE |
| cidr | TEXT | IPv6 CIDR block |
| last_modified | TEXT | Last modification date |

**Unique:** (country_id, cidr) — **Index:** `idx_country_ipv6_country`

## connectivity

| column | datatype | desc |
|--------|----------|------|
| id | SERIAL | Primary key |
| asn_id | INTEGER | FK → asn(id), ON DELETE CASCADE |
| provider_asn_id | INTEGER | FK → asn(id), ON DELETE CASCADE |
| customers | INTEGER | Customer link count |
| peers | INTEGER | Peer link count |
| unclassified | INTEGER | Unclassified link count |
| degree | INTEGER | Total connectivity degree |
| reach | DOUBLE PRECISION | Network reach metric |

**Unique:** (asn_id, provider_asn_id) — **Index:** `idx_connectivity_asn`

## api_request

| column | datatype | desc |
|--------|----------|------|
| id | SERIAL | Primary key |
| client_ip | TEXT | Client IP address |
| request_at | TIMESTAMPTZ | Request timestamp (default NOW()) |

**Index:** `idx_api_request_at`

## sync_state

| column | datatype | desc |
|--------|----------|------|
| id | INTEGER | Primary key (always 1) |
| last_sync_at | TIMESTAMPTZ | Last successful data sync timestamp |
