import { useEffect, useRef, useState } from 'react'
import { fetchIpInfo } from '../api/asIpClient'
import {
  gatherWebRtcCandidates,
  isWebRtcSupported,
  type GatheredCandidate,
} from '../lib/webrtcLeak'

type CheckState =
  | { status: 'idle' }
  | { status: 'unsupported' }
  | { status: 'running' }
  | {
      status: 'ready'
      candidates: GatheredCandidate[]
      callerIp: string | null
    }
  | { status: 'error'; message: string }

export function WebRtcPage() {
  const [checkState, setCheckState] = useState<CheckState>(() =>
    isWebRtcSupported() ? { status: 'idle' } : { status: 'unsupported' },
  )
  const abortRef = useRef<AbortController | null>(null)

  useEffect(() => {
    return () => {
      abortRef.current?.abort()
    }
  }, [])

  async function runCheck() {
    if (!isWebRtcSupported()) {
      setCheckState({ status: 'unsupported' })
      return
    }

    abortRef.current?.abort()
    const controller = new AbortController()
    abortRef.current = controller
    setCheckState({ status: 'running' })

    try {
      const [gatherResult, callerInfo] = await Promise.all([
        gatherWebRtcCandidates(controller.signal),
        fetchIpInfo().catch(() => null),
      ])

      if (controller.signal.aborted) {
        return
      }

      setCheckState({
        status: 'ready',
        candidates: gatherResult.candidates,
        callerIp: callerInfo?.ip ?? null,
      })
    } catch (error) {
      if (controller.signal.aborted) {
        return
      }
      const message =
        error instanceof Error ? error.message : 'Unable to gather WebRTC candidates'
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
          disabled={checkState.status === 'running' || checkState.status === 'unsupported'}
        >
          Run check
        </button>
      </div>

      {checkState.status === 'idle' && (
        <p className="status" role="status">
          Run a WebRTC ICE check to see local and public candidates.
        </p>
      )}

      {checkState.status === 'unsupported' && (
        <p className="status status-error" role="alert">
          WebRTC is not supported in this browser.
        </p>
      )}

      {checkState.status === 'running' && (
        <p className="status" role="status">
          Gathering ICE candidates…
        </p>
      )}

      {checkState.status === 'error' && (
        <p className="status status-error" role="alert">
          {checkState.message}
        </p>
      )}

      {checkState.status === 'ready' && (
        <WebRtcSummary
          candidates={checkState.candidates}
          callerIp={checkState.callerIp}
        />
      )}
    </div>
  )
}

function WebRtcSummary({
  candidates,
  callerIp,
}: {
  candidates: GatheredCandidate[]
  callerIp: string | null
}) {
  const publicIps = [
    ...new Set(
      candidates.filter((candidate) => !candidate.isPrivate).map((candidate) => candidate.ip),
    ),
  ]
  const matchesCaller =
    callerIp !== null && publicIps.some((ip) => ip === callerIp)

  return (
    <div className="dns-lookup-result">
      <dl className="ip-fields">
        <div className="field">
          <dt>Caller IP</dt>
          <dd>{callerIp ?? '—'}</dd>
        </div>
        <div className="field">
          <dt>Public WebRTC IPs</dt>
          <dd>{publicIps.length > 0 ? publicIps.join(', ') : 'None found'}</dd>
        </div>
        <div className="field">
          <dt>Match</dt>
          <dd className={matchesCaller ? 'leak-match' : 'leak-mismatch'}>
            {callerIp === null
              ? 'Caller IP unavailable'
              : matchesCaller
                ? 'Public candidate matches caller IP'
                : 'No public candidate matches caller IP'}
          </dd>
        </div>
      </dl>

      <div className="dns-grid">
        <h2 className="dns-grid-heading">ICE candidates</h2>
        {candidates.length === 0 ? (
          <p className="status dns-grid-empty">No candidates gathered.</p>
        ) : (
          <div className="dns-grid-scroll">
            <table className="dns-grid-table">
              <thead>
                <tr>
                  <th scope="col">Type</th>
                  <th scope="col">IP</th>
                  <th scope="col">Protocol</th>
                  <th scope="col">Scope</th>
                </tr>
              </thead>
              <tbody>
                {candidates.map((candidate) => (
                  <tr key={`${candidate.type}-${candidate.ip}-${candidate.protocol}`}>
                    <td>{candidate.type}</td>
                    <td>{candidate.ip}</td>
                    <td>{candidate.protocol || '—'}</td>
                    <td>{candidate.isPrivate ? 'private' : 'public'}</td>
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
