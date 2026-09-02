import { useEffect, useState } from 'react'
import { fetchIpInfo, type IpInfo } from '../api/asIpClient'
import { buildCountryFlagClassName } from '../lib/countryFlag'

type LoadState =
  | { status: 'loading' }
  | { status: 'ready'; info: IpInfo }
  | { status: 'error'; message: string }

export function HomePage() {
  const [loadState, setLoadState] = useState<LoadState>({ status: 'loading' })

  useEffect(() => {
    let isActive = true

    async function loadCallerIpInfo() {
      try {
        const info = await fetchIpInfo()
        if (isActive) {
          setLoadState({ status: 'ready', info })
        }
      } catch (error) {
        const message =
          error instanceof Error ? error.message : 'Unable to load IP information'
        if (isActive) {
          setLoadState({ status: 'error', message })
        }
      }
    }

    void loadCallerIpInfo()
    return () => {
      isActive = false
    }
  }, [])

  return (
    <div className="home">
      {loadState.status === 'loading' && (
        <p className="status" role="status">
          Looking up your IP…
        </p>
      )}

      {loadState.status === 'error' && (
        <p className="status status-error" role="alert">
          {loadState.message}
        </p>
      )}

      {loadState.status === 'ready' && (
        <CallerIpSummary info={loadState.info} />
      )}
    </div>
  )
}

function CallerIpSummary({ info }: { info: IpInfo }) {
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
