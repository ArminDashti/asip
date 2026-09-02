export type DohResolver = {
  id: string
  name: string
  endpoint: string
}

export type DnsLeakProbe = {
  resolver: DohResolver
  host: string
  recordType: 'A' | 'TXT'
  addresses: string[]
  error: string | null
}

export type DnsLeakCheckResult = {
  probes: DnsLeakProbe[]
  discoveredIps: string[]
}

/** Public DoH endpoints that accept DNS-JSON (`Accept: application/dns-json`). */
export const DOH_RESOLVERS: readonly DohResolver[] = [
  {
    id: 'cloudflare',
    name: 'Cloudflare',
    endpoint: 'https://cloudflare-dns.com/dns-query',
  },
  {
    id: 'google',
    name: 'Google',
    endpoint: 'https://dns.google/resolve',
  },
  {
    id: 'quad9',
    name: 'Quad9',
    endpoint: 'https://dns.quad9.net:5053/dns-query',
  },
] as const

/**
 * Hostnames whose DNS answers can reveal the querying client's public IP
 * (or an IP echo via TXT), used to compare against the API caller IP.
 */
const WHOAMI_PROBES: ReadonlyArray<{ host: string; recordType: 'A' | 'TXT' }> = [
  { host: 'whoami.cloudflare.com', recordType: 'TXT' },
  { host: 'o-o.myaddr.l.google.com', recordType: 'TXT' },
  { host: 'myip.opendns.com', recordType: 'A' },
]

type DnsJsonAnswer = {
  data?: string
}

type DnsJsonResponse = {
  Status?: number
  Answer?: DnsJsonAnswer[]
}

function isIpv4(value: string): boolean {
  const parts = value.split('.')
  if (parts.length !== 4) {
    return false
  }
  return parts.every((part) => {
    if (!/^\d{1,3}$/.test(part)) {
      return false
    }
    const n = Number(part)
    return n >= 0 && n <= 255
  })
}

function extractAddresses(recordType: 'A' | 'TXT', answers: DnsJsonAnswer[]): string[] {
  const found: string[] = []
  for (const answer of answers) {
    const raw = answer.data?.trim()
    if (!raw) {
      continue
    }
    if (recordType === 'A') {
      if (isIpv4(raw)) {
        found.push(raw)
      }
      continue
    }
    // TXT answers are often quoted; strip surrounding quotes and take IPv4 tokens.
    const unquoted = raw.replace(/^"+|"+$/g, '').trim()
    if (isIpv4(unquoted)) {
      found.push(unquoted)
      continue
    }
    for (const token of unquoted.split(/[\s,;]+/)) {
      if (isIpv4(token)) {
        found.push(token)
      }
    }
  }
  return [...new Set(found)]
}

export async function queryDoh(
  resolver: DohResolver,
  host: string,
  recordType: 'A' | 'TXT',
  signal?: AbortSignal,
): Promise<string[]> {
  const url = new URL(resolver.endpoint)
  url.searchParams.set('name', host)
  url.searchParams.set('type', recordType)

  const response = await fetch(url.toString(), {
    method: 'GET',
    headers: {
      Accept: 'application/dns-json',
    },
    signal,
  })

  if (!response.ok) {
    throw new Error(`${resolver.name} DoH HTTP ${response.status}`)
  }

  const body = (await response.json()) as DnsJsonResponse
  if (typeof body.Status === 'number' && body.Status !== 0) {
    throw new Error(`${resolver.name} DoH status ${body.Status}`)
  }

  return extractAddresses(recordType, body.Answer ?? [])
}

async function runProbe(
  resolver: DohResolver,
  host: string,
  recordType: 'A' | 'TXT',
  signal?: AbortSignal,
): Promise<DnsLeakProbe> {
  try {
    const addresses = await queryDoh(resolver, host, recordType, signal)
    return {
      resolver,
      host,
      recordType,
      addresses,
      error: null,
    }
  } catch (error) {
    if (signal?.aborted) {
      throw error
    }
    const message =
      error instanceof Error ? error.message : `Unable to query ${resolver.name}`
    return {
      resolver,
      host,
      recordType,
      addresses: [],
      error: message,
    }
  }
}

/**
 * Query multiple DoH resolvers for whoami-style records that may echo the
 * client public IP, then collect unique discovered addresses.
 */
export async function runDnsLeakCheck(
  signal?: AbortSignal,
): Promise<DnsLeakCheckResult> {
  const jobs: Array<Promise<DnsLeakProbe>> = []

  for (const resolver of DOH_RESOLVERS) {
    for (const probe of WHOAMI_PROBES) {
      // OpenDNS myip is only meaningful via resolvers that support it; still try all.
      jobs.push(runProbe(resolver, probe.host, probe.recordType, signal))
    }
  }

  const probes = await Promise.all(jobs)
  if (signal?.aborted) {
    throw new DOMException('Aborted', 'AbortError')
  }

  const discoveredIps = [
    ...new Set(probes.flatMap((probe) => probe.addresses)),
  ]

  return { probes, discoveredIps }
}
