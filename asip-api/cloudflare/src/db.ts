export type Env = {
  DB: D1Database
  CORS_ORIGINS: string
}

export type AsRecord = {
  id: number
  name: string
  category: string | null
  registered: string | null
  origin: string | null
  lastModified: string | null
  lastAnnounced: string | null
  countryCode: string | null
  countryName: string | null
  asnNumber: number
  prefixCount: number
  prefixesAgg: number
  largestPrefix: number | null
  totalAddresses: number
  providerAsns: number[]
}

const AS_RECORD_SELECT = `
SELECT
    a.id,
    a."as" AS as_name,
    cat.category,
    a.registered,
    orig.origin,
    a.last_modified,
    a.last_announced,
    c.country_code,
    c.country,
    a.asn,
    COALESCE(s.prefixes, 0) AS prefixes,
    COALESCE(s.prefixes_aggregated, 0) AS prefixes_aggregated,
    s.largest_prefix,
    COALESCE(s.total_addresses, 0) AS total_addresses,
    COALESCE((
        SELECT GROUP_CONCAT(provider.asn)
        FROM connectivity conn
        JOIN asn provider ON provider.id = conn.provider_asn_id
        WHERE conn.asn_id = a.id
    ), '') AS provider_asns
FROM asn a
LEFT JOIN country c ON c.id = a.country_id
LEFT JOIN category cat ON cat.id = a.category_id
LEFT JOIN origin orig ON orig.id = a.origin_id
LEFT JOIN ipv4_stat s ON s.asn_id = a.id
`

function parseProviderAsns(value: string | null): number[] {
  if (!value) return []
  return value
    .split(',')
    .map((part) => Number.parseInt(part.trim(), 10))
    .filter((n) => Number.isFinite(n))
}

function mapRow(row: Record<string, unknown>): AsRecord {
  return {
    id: Number(row.id),
    name: String(row.as_name ?? ''),
    category: (row.category as string | null) ?? null,
    registered: (row.registered as string | null) ?? null,
    origin: (row.origin as string | null) ?? null,
    lastModified: (row.last_modified as string | null) ?? null,
    lastAnnounced: (row.last_announced as string | null) ?? null,
    countryCode: (row.country_code as string | null) ?? null,
    countryName: (row.country as string | null) ?? null,
    asnNumber: Number(row.asn),
    prefixCount: Number(row.prefixes ?? 0),
    prefixesAgg: Number(row.prefixes_aggregated ?? 0),
    largestPrefix:
      row.largest_prefix === null || row.largest_prefix === undefined
        ? null
        : Number(row.largest_prefix),
    totalAddresses: Number(row.total_addresses ?? 0),
    providerAsns: parseProviderAsns((row.provider_asns as string | null) ?? null),
  }
}

export function ipv4ToInt(ip: string): number | null {
  const parts = ip.split('.')
  if (parts.length !== 4) return null
  let value = 0
  for (const part of parts) {
    const n = Number.parseInt(part, 10)
    if (!Number.isInteger(n) || n < 0 || n > 255) return null
    value = (value << 8) + n
  }
  return value >>> 0
}

export async function logRequest(db: D1Database, clientIp: string): Promise<void> {
  await db
    .prepare('INSERT INTO api_request (client_ip) VALUES (?)')
    .bind(clientIp)
    .run()
}

export async function findByIp(db: D1Database, ip: string): Promise<AsRecord | null> {
  const ipInt = ipv4ToInt(ip)
  if (ipInt === null) return null
  const row = await db
    .prepare(
      `${AS_RECORD_SELECT}
WHERE EXISTS (
  SELECT 1 FROM asn_ipv4 p
  WHERE p.asn_id = a.id AND ?1 >= p.start_ip AND ?1 <= p.end_ip
)
ORDER BY s.largest_prefix IS NULL, s.largest_prefix DESC, a.asn ASC
LIMIT 1`,
    )
    .bind(ipInt)
    .first()
  return row ? mapRow(row as Record<string, unknown>) : null
}

export async function findCountryByIp(
  db: D1Database,
  ip: string,
): Promise<{ code: string; name: string } | null> {
  const ipInt = ipv4ToInt(ip)
  if (ipInt === null) return null
  const row = await db
    .prepare(
      `SELECT c.country_code AS code, c.country AS name
       FROM country_ipv4 p
       JOIN country c ON c.id = p.country_id
       WHERE ?1 >= p.start_ip AND ?1 <= p.end_ip
       ORDER BY (p.end_ip - p.start_ip) ASC
       LIMIT 1`,
    )
    .bind(ipInt)
    .first<{ code: string; name: string }>()
  return row ?? null
}

export async function findByAsnNumber(db: D1Database, asn: number): Promise<AsRecord | null> {
  const row = await db
    .prepare(`${AS_RECORD_SELECT} WHERE a.asn = ?1 LIMIT 1`)
    .bind(asn)
    .first()
  return row ? mapRow(row as Record<string, unknown>) : null
}

export async function findByAsIdentifier(
  db: D1Database,
  identifier: string,
): Promise<AsRecord | null> {
  const normalized = identifier.trim()
  if (!normalized) return null
  const asNumber = Number.parseInt(normalized.replace(/^AS/i, ''), 10)
  if (Number.isFinite(asNumber)) {
    const byNumber = await findByAsnNumber(db, asNumber)
    if (byNumber) return byNumber
  }
  const row = await db
    .prepare(
      `${AS_RECORD_SELECT}
WHERE LOWER(a."as") LIKE LOWER('%' || ?1 || '%')
ORDER BY a.asn ASC
LIMIT 1`,
    )
    .bind(normalized)
    .first()
  return row ? mapRow(row as Record<string, unknown>) : null
}

export async function listByCountry(db: D1Database, countryCode: string): Promise<AsRecord[]> {
  const { results } = await db
    .prepare(
      `${AS_RECORD_SELECT}
WHERE LOWER(c.country_code) = LOWER(?1)
ORDER BY a.asn ASC`,
    )
    .bind(countryCode)
    .all()
  return (results ?? []).map((row) => mapRow(row as Record<string, unknown>))
}

export async function listCountries(
  db: D1Database,
): Promise<Array<{ code: string; name: string }>> {
  const { results } = await db
    .prepare(
      `SELECT country_code AS code, country AS name FROM country ORDER BY country_code`,
    )
    .all<{ code: string; name: string }>()
  return results ?? []
}

export async function listAsns(
  db: D1Database,
): Promise<Array<{ number: number; name: string }>> {
  const { results } = await db
    .prepare(`SELECT asn AS number, "as" AS name FROM asn ORDER BY asn`)
    .all<{ number: number; name: string }>()
  return results ?? []
}

export function toAsnResponse(record: AsRecord) {
  return {
    asn: record.asnNumber,
    metadata: {
      handle: record.name,
      description: record.name,
      countryCode: record.countryCode ?? '',
      country: record.countryName ?? '',
      origin: record.origin,
      category: record.category,
      registered: record.registered,
      lastModified: record.lastModified,
    },
    stats: {
      ipv4: {
        prefixes: record.prefixCount,
        prefixesAggregated: record.prefixesAgg,
        largestPrefix: record.largestPrefix,
        totalAddresses: record.totalAddresses,
      },
      connectivity: {
        providers: record.providerAsns.length,
        providerAsns: record.providerAsns,
      },
    },
    lastAnnounced: record.lastAnnounced,
  }
}

export function toAsSummary(record: AsRecord) {
  return {
    id: record.id,
    handle: record.name,
    organization: record.name,
    countryCode: record.countryCode,
    country: record.countryName,
  }
}
