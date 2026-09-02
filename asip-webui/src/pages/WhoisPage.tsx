import { useState, type FormEvent } from 'react'
import {
  fetchIpWhois,
  fetchIpWhoisByAddress,
  type WhoisResult,
} from '../api/asIpClient'

type LookupState =
  | { status: 'idle' }
  | { status: 'loading' }
  | { status: 'ready'; result: WhoisResult }
  | { status: 'error'; message: string }

export function WhoisPage() {
  const [ipAddress, setIpAddress] = useState('')
  const [lookupState, setLookupState] = useState<LookupState>({ status: 'idle' })

  async function runLookup(address: string | null) {
    setLookupState({ status: 'loading' })
    try {
      const result =
        address === null ? await fetchIpWhois() : await fetchIpWhoisByAddress(address)
      setLookupState({ status: 'ready', result })
    } catch (error) {
      const message =
        error instanceof Error ? error.message : 'Unable to look up whois for that IP'
      setLookupState({ status: 'error', message })
    }
  }

  async function handleLookup(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const trimmed = ipAddress.trim()
    if (!trimmed) {
      setLookupState({ status: 'error', message: 'Enter an IP address' })
      return
    }
    await runLookup(trimmed)
  }

  async function handleMyIp() {
    setIpAddress('')
    await runLookup(null)
  }

  return (
    <div className="home">
      <form className="ip-lookup-form" onSubmit={handleLookup}>
        <label className="ip-lookup-label" htmlFor="whois-ip-address">
          IP address
        </label>
        <div className="ip-lookup-row">
          <input
            id="whois-ip-address"
            className="ip-lookup-input"
            type="text"
            name="ip"
            value={ipAddress}
            onChange={(event) => setIpAddress(event.target.value)}
            placeholder="8.8.8.8"
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

      <div className="page-actions">
        <button
          className="ip-lookup-submit"
          type="button"
          onClick={() => void handleMyIp()}
          disabled={lookupState.status === 'loading'}
        >
          My IP
        </button>
      </div>

      {lookupState.status === 'loading' && (
        <p className="status" role="status">
          Looking up whois…
        </p>
      )}

      {lookupState.status === 'error' && (
        <p className="status status-error" role="alert">
          {lookupState.message}
        </p>
      )}

      {lookupState.status === 'ready' && <WhoisSummary result={lookupState.result} />}
    </div>
  )
}

function WhoisSummary({ result }: { result: WhoisResult }) {
  const statusText = result.status.length > 0 ? result.status.join(', ') : '—'
  const entitiesText = result.entities.length > 0 ? result.entities.join(', ') : '—'
  const rangeText =
    result.startAddress && result.endAddress
      ? `${result.startAddress} – ${result.endAddress}`
      : '—'

  return (
    <div className="dns-lookup-result">
      <dl className="ip-fields">
        <div className="field">
          <dt>IP</dt>
          <dd>{result.ip || '—'}</dd>
        </div>
        <div className="field">
          <dt>Handle</dt>
          <dd>{result.handle || '—'}</dd>
        </div>
        <div className="field">
          <dt>Name</dt>
          <dd>{result.name || '—'}</dd>
        </div>
        <div className="field">
          <dt>Type</dt>
          <dd>{result.type || '—'}</dd>
        </div>
        <div className="field">
          <dt>Country</dt>
          <dd>{result.country || '—'}</dd>
        </div>
        <div className="field">
          <dt>Range</dt>
          <dd>{rangeText}</dd>
        </div>
        <div className="field">
          <dt>Status</dt>
          <dd>{statusText}</dd>
        </div>
        <div className="field">
          <dt>Entities</dt>
          <dd>{entitiesText}</dd>
        </div>
        <div className="field">
          <dt>RDAP</dt>
          <dd>
            {result.rdapUrl ? (
              <a href={result.rdapUrl} target="_blank" rel="noopener noreferrer">
                {result.rdapUrl}
              </a>
            ) : (
              '—'
            )}
          </dd>
        </div>
      </dl>

      {result.events.length > 0 && (
        <div className="dns-grid">
          <h2 className="dns-grid-heading">Events</h2>
          <div className="dns-grid-scroll">
            <table className="dns-grid-table">
              <thead>
                <tr>
                  <th scope="col">Action</th>
                  <th scope="col">Date</th>
                </tr>
              </thead>
              <tbody>
                {result.events.map((event) => (
                  <tr key={`${event.action}-${event.date}`}>
                    <td>{event.action || '—'}</td>
                    <td>{event.date || '—'}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {result.remarks.length > 0 && (
        <pre className="whois-remarks">{result.remarks.join('\n')}</pre>
      )}
    </div>
  )
}
