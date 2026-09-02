import { useEffect, useRef, useState } from 'react'
import { fetchDnsLookup, fetchIpInfo } from '../api/asIpClient'
import {
  runDnsLeakCheck,
  type DnsLeakProbe,
} from '../lib/dnsLeak'

type CheckState =
  | { status: 'idle' }
  | { status: 'running' }
  | {
      status: 'ready'
      callerIp: string | null
      discoveredIps: string[]
      probes: DnsLeakProbe[]
      serverLookupIps: string[]
      serverLookupError: string | null
    }
  | { status: 'error'; message: string }

const SERVER_COMPARE_DOMAIN = 'whoami.cloudflare.com'

function displayOrDash(value: string | null | undefined): string {
  if (value === undefined || value === null) {
    return '—'
  }
  const text = value.trim()
  return text.length > 0 ? text : '—'
}

export function DnsLeakPage() {
  const [checkState, setCheckState] = useState<CheckState>({ status: 'idle' })
  const abortRef = useRef<AbortController | null>(null)

  useEffect(() => {
    return () => {
      abortRef.current?.abort()
    }
  }, [])

  async function runCheck() {
    abortRef.current?.abort()
    const controller = new AbortController()
    abortRef.current = controller
    setCheckState({ status: 'running' })

    try {
      const [leakResult, callerInfo, serverLookup] = await Promise.all([
        runDnsLeakCheck(controller.signal),
        fetchIpInfo().catch(() => null),
        fetchDnsLookup(SERVER_COMPARE_DOMAIN)
          .then((result) => ({
            ips: [...result.a, ...result.aaaa],
            error: null as string | null,
          }))
          .catch((error: unknown) => ({
            ips: [] as string[],
            error:
              error instanceof Error
                ? error.message
                : 'Server DNS lookup failed',
          })),
      ])

      if (controller.signal.aborted) {
        return
      }

      setCheckState({
        status: 'ready',
        callerIp: callerInfo?.ip ?? null,
        discoveredIps: leakResult.discoveredIps,
        probes: leakResult.probes,
        serverLookupIps: serverLookup.ips,
        serverLookupError: serverLookup.error,
      })
    } catch (error) {
      if (controller.signal.aborted) {
        return
      }
      const message =
        error instanceof Error ? error.message : 'Unable to run DNS leak check'
      setCheckState({ status: 'error', message })
    }
  }

  return (
    <div className="home">
      <div className="page-actions">
        <button
          className="ip-lookup-submit"
          type="button"
          onClick={() => void runCheck()}
          disabled={checkState.status === 'running'}
        >
          Run check
        </button>
      </div>

      {checkState.status === 'idle' && (
        <p className="status" role="status">
          Run a DNS leak check via browser DoH resolvers and compare discovered
          IPs to your caller IP.
        </p>
      )}

      {checkState.status === 'running' && (
        <p className="status" role="status">
          Querying DoH resolvers…
        </p>
      )}

      {checkState.status === 'error' && (
        <p className="status status-error" role="alert">
          {checkState.message}
        </p>
      )}

      {checkState.status === 'ready' && (
        <DnsLeakSummary
          callerIp={checkState.callerIp}
          discoveredIps={checkState.discoveredIps}
          probes={checkState.probes}
          serverLookupIps={checkState.serverLookupIps}
          serverLookupError={checkState.serverLookupError}
        />
      )}
    </div>
  )
}

function DnsLeakSummary({
  callerIp,
  discoveredIps,
  probes,
  serverLookupIps,
  serverLookupError,
}: {
  callerIp: string | null
  discoveredIps: string[]
  probes: DnsLeakProbe[]
  serverLookupIps: string[]
  serverLookupError: string | null
}) {
  const matchesCaller =
    callerIp !== null && discoveredIps.some((ip) => ip === callerIp)
  const hasDiscoveredIps = discoveredIps.length > 0

  return (
    <div className="dns-lookup-result">
      <dl className="ip-fields">
        <div className="field">
          <dt>Caller IP</dt>
          <dd>{displayOrDash(callerIp)}</dd>
        </div>
        <div className="field">
          <dt>DoH-discovered IPs</dt>
          <dd>
            {hasDiscoveredIps ? discoveredIps.join(', ') : 'None found'}
          </dd>
        </div>
        <div className="field">
          <dt>API DNS ({SERVER_COMPARE_DOMAIN})</dt>
          <dd>
            {serverLookupError
              ? serverLookupError
              : serverLookupIps.length > 0
                ? serverLookupIps.join(', ')
                : 'No addresses'}
          </dd>
        </div>
        <div className="field">
          <dt>Match</dt>
          <dd
            className={
              !hasDiscoveredIps || callerIp === null
                ? undefined
                : matchesCaller
                  ? 'leak-match'
                  : 'leak-mismatch'
            }
          >
            {callerIp === null
              ? 'Caller IP unavailable'
              : !hasDiscoveredIps
                ? 'No client IPs discovered via DoH whoami probes'
                : matchesCaller
                  ? 'Discovered IP matches caller IP'
                  : 'Discovered IP differs from caller IP (possible DNS path mismatch)'}
          </dd>
        </div>
      </dl>

      <div className="dns-grid">
        <h2 className="dns-grid-heading">DoH probes</h2>
        {probes.length === 0 ? (
          <p className="status dns-grid-empty">No probes ran.</p>
        ) : (
          <div className="dns-grid-scroll">
            <table className="dns-grid-table">
              <thead>
                <tr>
                  <th scope="col">Resolver</th>
                  <th scope="col">Host</th>
                  <th scope="col">Type</th>
                  <th scope="col">Addresses</th>
                  <th scope="col">Status</th>
                </tr>
              </thead>
              <tbody>
                {probes.map((probe) => (
                  <tr
                    key={`${probe.resolver.id}-${probe.host}-${probe.recordType}`}
                  >
                    <td>{probe.resolver.name}</td>
                    <td>{probe.host}</td>
                    <td>{probe.recordType}</td>
                    <td>
                      {probe.addresses.length > 0
                        ? probe.addresses.join(', ')
                        : '—'}
                    </td>
                    <td>{probe.error ?? 'ok'}</td>
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
