import { useEffect, useState } from 'react'
import { fetchHttpHeaders, type HttpHeadersResult } from '../api/asIpClient'

type LoadState =
  | { status: 'loading' }
  | { status: 'ready'; result: HttpHeadersResult }
  | { status: 'error'; message: string }

export function HeadersPage() {
  const [loadState, setLoadState] = useState<LoadState>({ status: 'loading' })

  useEffect(() => {
    let isActive = true

    async function loadHeaders() {
      try {
        const result = await fetchHttpHeaders()
        if (isActive) {
          setLoadState({ status: 'ready', result })
        }
      } catch (error) {
        const message =
          error instanceof Error ? error.message : 'Unable to load HTTP headers'
        if (isActive) {
          setLoadState({ status: 'error', message })
        }
      }
    }

    void loadHeaders()
    return () => {
      isActive = false
    }
  }, [])

  return (
    <div className="home">
      {loadState.status === 'loading' && (
        <p className="status" role="status">
          Reading request headers…
        </p>
      )}

      {loadState.status === 'error' && (
        <p className="status status-error" role="alert">
          {loadState.message}
        </p>
      )}

      {loadState.status === 'ready' && <HeadersSummary result={loadState.result} />}
    </div>
  )
}

function HeadersSummary({ result }: { result: HttpHeadersResult }) {
  return (
    <div className="dns-lookup-result">
      <dl className="ip-fields">
        <div className="field">
          <dt>IP</dt>
          <dd>{result.ip || '—'}</dd>
        </div>
      </dl>

      <div className="dns-grid">
        <h2 className="dns-grid-heading">Headers seen by API</h2>
        {result.headers.length === 0 ? (
          <p className="status dns-grid-empty">No headers returned.</p>
        ) : (
          <div className="dns-grid-scroll">
            <table className="dns-grid-table">
              <thead>
                <tr>
                  <th scope="col">Name</th>
                  <th scope="col">Value</th>
                </tr>
              </thead>
              <tbody>
                {result.headers.map((header) => (
                  <tr key={header.name}>
                    <td>{header.name}</td>
                    <td>{header.value}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  )
}
