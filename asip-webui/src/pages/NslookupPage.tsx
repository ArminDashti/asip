import { useEffect, useState, type FormEvent } from 'react'
import {
  fetchDnsLookup,
  type DnsAddress,
  type DnsLookupResult,
} from '../api/asIpClient'
import { buildCountryFlagClassName } from '../lib/countryFlag'

type LookupState =
  | { status: 'idle' }
  | { status: 'loading' }
  | { status: 'ready'; result: DnsLookupResult }
  | { status: 'error'; message: string }

type NsRow = {
  ns: string
  ip: string
  as: string
  asn: string
  country: string
}

type NsResolveState =
  | { status: 'loading' }
  | { status: 'ready'; rows: NsRow[] }

function displayOrDash(value: string | number | undefined | null): string {
  if (value === undefined || value === null) {
    return '—'
  }
  const text = String(value).trim()
  return text.length > 0 ? text : '—'
}

function joinRecordList(values: string[]): string {
  return values.length > 0 ? values.join(', ') : '—'
}

function mapDnsLookupToNsRow(nameserver: string, lookup: DnsLookupResult): NsRow {
  const primaryAddress = lookup.addresses[0]
  const ip = primaryAddress?.ip || lookup.a[0]
  const asName = primaryAddress?.as ?? lookup.as
  const asn = primaryAddress?.asn ?? lookup.asn
  const country = primaryAddress?.country ?? lookup.country

  return {
    ns: nameserver,
    ip: displayOrDash(ip),
    as: displayOrDash(asName),
    asn: displayOrDash(asn),
    country: displayOrDash(country),
  }
}

export function NslookupPage() {
  const [domainName, setDomainName] = useState('')
  const [lookupState, setLookupState] = useState<LookupState>({ status: 'idle' })

  async function handleLookup(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const trimmed = domainName.trim()
    if (!trimmed) {
      setLookupState({ status: 'error', message: 'Enter a domain name' })
      return
    }

    setLookupState({ status: 'loading' })
    try {
      const result = await fetchDnsLookup(trimmed)
      setLookupState({ status: 'ready', result })
    } catch (error) {
      const message =
        error instanceof Error ? error.message : 'Unable to look up that domain'
      setLookupState({ status: 'error', message })
    }
  }

  return (
    <div className="home">
      <form className="ip-lookup-form" onSubmit={handleLookup}>
        <label className="ip-lookup-label" htmlFor="domain-name">
          Domain
        </label>
        <div className="ip-lookup-row">
          <input
            id="domain-name"
            className="ip-lookup-input"
            type="text"
            name="domain"
            value={domainName}
            onChange={(event) => setDomainName(event.target.value)}
            placeholder="example.com"
            autoComplete="off"
            spellCheck={false}
            disabled={lookupState.status === 'loading'}
          />
          <button
            className="ip-lookup-submit"
            type="submit"
            disabled={lookupState.status === 'loading'}
          >
            Lookup
          </button>
        </div>
      </form>

      {lookupState.status === 'loading' && (
        <p className="status" role="status">
          Looking up domain…
        </p>
      )}

      {lookupState.status === 'error' && (
        <p className="status status-error" role="alert">
          {lookupState.message}
        </p>
      )}

      {lookupState.status === 'ready' && (
        <DnsLookupSummary result={lookupState.result} />
      )}
    </div>
  )
}

function DnsLookupSummary({ result }: { result: DnsLookupResult }) {
  const flagClassName = buildCountryFlagClassName(result.country)

  return (
    <div className="dns-lookup-result">
      <dl className="ip-fields">
        {flagClassName && (
          <div className="flag-row">
            <span className={`country-flag ${flagClassName}`} aria-hidden="true" />
          </div>
        )}
        <div className="field">
          <dt>IP</dt>
          <dd>{joinRecordList(result.a)}</dd>
        </div>
        <div className="field">
          <dt>ASN</dt>
          <dd>{displayOrDash(result.asn)}</dd>
        </div>
        <div className="field">
          <dt>AS</dt>
          <dd>{displayOrDash(result.as)}</dd>
        </div>
        <div className="field">
          <dt>Country</dt>
          <dd>{displayOrDash(result.country)}</dd>
        </div>
      </dl>

      <DnsAddressesTable addresses={result.addresses} />
      <DnsNsTable nameservers={result.ns} />
    </div>
  )
}

function DnsAddressesTable({ addresses }: { addresses: DnsAddress[] }) {
  if (addresses.length === 0) {
    return (
      <p className="status dns-grid-empty" role="status">
        No addresses returned
      </p>
    )
  }

  return (
    <div className="dns-grid">
      <div className="dns-grid-scroll">
        <table className="dns-grid-table">
          <thead>
            <tr>
              <th scope="col">IP</th>
              <th scope="col">ASN</th>
              <th scope="col">AS</th>
              <th scope="col">Country</th>
            </tr>
          </thead>
          <tbody>
            {addresses.map((address) => (
              <DnsAddressRow key={address.ip} address={address} />
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}

function DnsAddressRow({ address }: { address: DnsAddress }) {
  const flagClassName = buildCountryFlagClassName(address.country)

  return (
    <tr>
      <td>{address.ip}</td>
      <td>{displayOrDash(address.asn)}</td>
      <td>{displayOrDash(address.as)}</td>
      <td className="dns-country-cell">
        {flagClassName && (
          <span
            className={`country-flag country-flag-inline ${flagClassName}`}
            aria-hidden="true"
          />
        )}
        <span>{displayOrDash(address.country)}</span>
      </td>
    </tr>
  )
}

function DnsNsTable({ nameservers }: { nameservers: string[] }) {
  const [resolveState, setResolveState] = useState<NsResolveState>({
    status: 'loading',
  })

  const nameserverKey = nameservers.join('\n')

  useEffect(() => {
    let isActive = true
    const nameserverList = nameserverKey.length > 0 ? nameserverKey.split('\n') : []

    if (nameserverList.length === 0) {
      setResolveState({ status: 'ready', rows: [] })
      return () => {
        isActive = false
      }
    }

    setResolveState({ status: 'loading' })

    async function resolveNameservers() {
      const settled = await Promise.allSettled(
        nameserverList.map((nameserver) => fetchDnsLookup(nameserver)),
      )

      if (!isActive) {
        return
      }

      const rows = settled.map((outcome, index) => {
        const nameserver = nameserverList[index]
        if (outcome.status === 'fulfilled') {
          return mapDnsLookupToNsRow(nameserver, outcome.value)
        }
        return {
          ns: nameserver,
          ip: '—',
          as: '—',
          asn: '—',
          country: '—',
        }
      })

      setResolveState({ status: 'ready', rows })
    }

    void resolveNameservers()

    return () => {
      isActive = false
    }
  }, [nameserverKey])

  if (nameservers.length === 0) {
    return (
      <p className="status dns-grid-empty" role="status">
        No nameservers returned
      </p>
    )
  }

  return (
    <div className="dns-grid">
      {resolveState.status === 'loading' && (
        <p className="status dns-grid-empty" role="status">
          Looking up nameservers…
        </p>
      )}
      {resolveState.status === 'ready' && (
        <div className="dns-grid-scroll">
          <table className="dns-grid-table">
            <thead>
              <tr>
                <th scope="col">NS</th>
                <th scope="col">IP</th>
                <th scope="col">ASN</th>
                <th scope="col">AS</th>
                <th scope="col">Country</th>
              </tr>
            </thead>
            <tbody>
              {resolveState.rows.map((row) => (
                <DnsNsRow key={row.ns} row={row} />
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}

function DnsNsRow({ row }: { row: NsRow }) {
  const flagClassName = buildCountryFlagClassName(
    row.country === '—' ? '' : row.country,
  )

  return (
    <tr>
      <td>{row.ns}</td>
      <td>{row.ip}</td>
      <td>{row.asn}</td>
      <td>{row.as}</td>
      <td className="dns-country-cell">
        {flagClassName && (
          <span
            className={`country-flag country-flag-inline ${flagClassName}`}
            aria-hidden="true"
          />
        )}
        <span>{row.country}</span>
      </td>
    </tr>
  )
}
