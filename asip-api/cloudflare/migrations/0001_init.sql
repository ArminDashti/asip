-- AS-IP schema for Cloudflare D1 (SQLite)

CREATE TABLE IF NOT EXISTS country (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    country      TEXT NOT NULL,
    country_code TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS origin (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    origin      TEXT NOT NULL UNIQUE,
    description TEXT
);

CREATE TABLE IF NOT EXISTS category (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    category    TEXT NOT NULL UNIQUE,
    description TEXT
);

CREATE TABLE IF NOT EXISTS role (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    role        TEXT NOT NULL UNIQUE,
    description TEXT
);

CREATE TABLE IF NOT EXISTS asn (
    id                     INTEGER PRIMARY KEY AUTOINCREMENT,
    asn                    INTEGER NOT NULL UNIQUE,
    "as"                   TEXT NOT NULL,
    country_id             INTEGER REFERENCES country (id) ON DELETE SET NULL,
    origin_id              INTEGER REFERENCES origin (id) ON DELETE SET NULL,
    category_id            INTEGER REFERENCES category (id) ON DELETE SET NULL,
    network_role_id        INTEGER REFERENCES role (id) ON DELETE SET NULL,
    registered             TEXT,
    last_modified          TEXT,
    last_announced         TEXT,
    prefixes_last_modified TEXT
);

CREATE INDEX IF NOT EXISTS idx_asn_asn ON asn (asn);
CREATE INDEX IF NOT EXISTS idx_asn_country ON asn (country_id);

CREATE TABLE IF NOT EXISTS ipv4_stat (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
    asn_id              INTEGER NOT NULL UNIQUE REFERENCES asn (id) ON DELETE CASCADE,
    prefixes            INTEGER NOT NULL DEFAULT 0,
    prefixes_aggregated INTEGER NOT NULL DEFAULT 0,
    largest_prefix      INTEGER,
    total_addresses     INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS ipv6_stat (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
    asn_id              INTEGER NOT NULL UNIQUE REFERENCES asn (id) ON DELETE CASCADE,
    prefixes            INTEGER NOT NULL DEFAULT 0,
    prefixes_aggregated INTEGER NOT NULL DEFAULT 0,
    largest_prefix      INTEGER,
    total_addresses     INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS asn_ipv4 (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    asn_id        INTEGER NOT NULL REFERENCES asn (id) ON DELETE CASCADE,
    cidr          TEXT NOT NULL,
    last_modified TEXT,
    start_ip      INTEGER NOT NULL,
    end_ip        INTEGER NOT NULL,
    UNIQUE (asn_id, cidr)
);

CREATE INDEX IF NOT EXISTS idx_asn_ipv4_asn ON asn_ipv4 (asn_id);
CREATE INDEX IF NOT EXISTS idx_asn_ipv4_range ON asn_ipv4 (start_ip, end_ip);

CREATE TABLE IF NOT EXISTS asn_ipv6 (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    asn_id        INTEGER NOT NULL REFERENCES asn (id) ON DELETE CASCADE,
    cidr          TEXT NOT NULL,
    last_modified TEXT,
    UNIQUE (asn_id, cidr)
);

CREATE INDEX IF NOT EXISTS idx_asn_ipv6_asn ON asn_ipv6 (asn_id);

CREATE TABLE IF NOT EXISTS country_ipv4 (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    country_id    INTEGER NOT NULL REFERENCES country (id) ON DELETE CASCADE,
    cidr          TEXT NOT NULL,
    last_modified TEXT,
    start_ip      INTEGER NOT NULL,
    end_ip        INTEGER NOT NULL,
    UNIQUE (country_id, cidr)
);

CREATE INDEX IF NOT EXISTS idx_country_ipv4_country ON country_ipv4 (country_id);
CREATE INDEX IF NOT EXISTS idx_country_ipv4_range ON country_ipv4 (start_ip, end_ip);

CREATE TABLE IF NOT EXISTS country_ipv6 (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    country_id    INTEGER NOT NULL REFERENCES country (id) ON DELETE CASCADE,
    cidr          TEXT NOT NULL,
    last_modified TEXT,
    UNIQUE (country_id, cidr)
);

CREATE INDEX IF NOT EXISTS idx_country_ipv6_country ON country_ipv6 (country_id);

CREATE TABLE IF NOT EXISTS connectivity (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    asn_id          INTEGER NOT NULL REFERENCES asn (id) ON DELETE CASCADE,
    provider_asn_id INTEGER NOT NULL REFERENCES asn (id) ON DELETE CASCADE,
    customers       INTEGER NOT NULL DEFAULT 0,
    peers           INTEGER NOT NULL DEFAULT 0,
    unclassified    INTEGER NOT NULL DEFAULT 0,
    degree          INTEGER,
    reach           REAL,
    UNIQUE (asn_id, provider_asn_id)
);

CREATE INDEX IF NOT EXISTS idx_connectivity_asn ON connectivity (asn_id);

CREATE TABLE IF NOT EXISTS api_request (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    client_ip  TEXT NOT NULL,
    request_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_api_request_at ON api_request (request_at);

CREATE TABLE IF NOT EXISTS sync_state (
    id           INTEGER PRIMARY KEY CHECK (id = 1),
    last_sync_at TEXT
);
