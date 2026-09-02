import { useState, type FormEvent } from 'react'
import { fetchIpInfoByAddress, type IpInfo } from '../api/asIpClient'
import { buildCountryFlagClassName } from '../lib/countryFlag'

type LookupState =
  | { status: 'idle' }
  | { status: 'loading' }
  | { status: 'ready'; info: IpInfo }
  | { status: 'error'; message: string }

export function IpPage() {
  const [ipAddress, setIpAddress] = useState('')
  const [lookupState, setLookupState] = useState<LookupState>({ status: 'idle' })

  async function handleLookup(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const trimmed = ipAddress.trim()
    if (!trimmed) {
      setLookupState({ status: 'error', message: 'Enter an IP address' })
      return
    }

    setLookupState({ status: 'loading' })
    try {
      const info = await fetchIpInfoByAddress(trimmed)
      setLookupState({ status: 'ready', info })
    } catch (error) {
      const message =
        error instanceof Error ? error.message : 'Unable to look up that IP'
      setLookupState({ status: 'error', message })
    }
  }

  return (
    <div className="home">
      <form className="ip-lookup-form" onSubmit={handleLookup}>
        <label className="ip-lookup-label" htmlFor="ip-address">
          IP address
        </label>
        <div className="ip-lookup-row">
          <input
            id="ip-address"
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

      {lookupState.status === 'loading' && (
        <p className="status" role="status">
          Looking up IP…
        </p>
      )}

      {lookupState.status === 'error' && (
        <p className="status status-error" role="alert">
          {lookupState.message}
        </p>
      )}

      {lookupState.status === 'ready' && (
        <LookedUpIpSummary info={lookupState.info} />
      )}
    </div>
  )
}

function LookedUpIpSummary({ info }: { info: IpInfo }) {
  const flagClassName = buildCountryFlagClassName(info.country)

  return (
    <dl className="ip-fields">
      {flagClassName && (
        <div className="flag-row">
          <span className={`country-flag ${flagClassName}`} aria-hidden="true" />
        </div>
      )}
      <div className="field">
        <dt>IP</dt>
        <dd>{info.ip}</dd>
      </div>
      <div className="field">
        <dt>ASN</dt>
        <dd>{info.asn}</dd>
      </div>
      <div className="field">
        <dt>AS</dt>
        <dd>{info.as}</dd>
      </div>
      <div className="field">
        <dt>Country</dt>
        <dd>{info.country}</dd>
      </div>
    </dl>
  )
}
