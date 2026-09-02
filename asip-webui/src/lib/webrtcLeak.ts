export type IceCandidateType = 'host' | 'srflx' | 'relay' | 'prflx' | 'unknown'

export type GatheredCandidate = {
  ip: string
  type: IceCandidateType
  protocol: string
  raw: string
  isPrivate: boolean
}

export type WebRtcGatherResult = {
  candidates: GatheredCandidate[]
}

const STUN_URLS = ['stun:stun.l.google.com:19302', 'stun:stun1.l.google.com:19302']

export function isWebRtcSupported(): boolean {
  return typeof RTCPeerConnection !== 'undefined'
}

export function isPrivateIp(ip: string): boolean {
  const trimmed = ip.trim().toLowerCase()
  if (trimmed === '::1' || trimmed === '0.0.0.0') {
    return true
  }
  if (trimmed.startsWith('fc') || trimmed.startsWith('fd') || trimmed.startsWith('fe80:')) {
    return true
  }

  const ipv4 = trimmed.match(/^(\d{1,3})\.(\d{1,3})\.(\d{1,3})\.(\d{1,3})$/)
  if (!ipv4) {
    return false
  }

  const [a, b] = [Number(ipv4[1]), Number(ipv4[2])]
  if (a === 10 || a === 127) {
    return true
  }
  if (a === 192 && b === 168) {
    return true
  }
  if (a === 172 && b >= 16 && b <= 31) {
    return true
  }
  if (a === 169 && b === 254) {
    return true
  }
  return false
}

function parseCandidate(candidate: string): GatheredCandidate | null {
  const tokens = candidate.replace(/^candidate:/i, '').trim().split(/\s+/)
  if (tokens.length < 8) {
    return null
  }

  const protocol = tokens[2]?.toLowerCase() ?? ''
  const ip = tokens[4]
  const typIndex = tokens.indexOf('typ')
  const typeToken = typIndex >= 0 ? tokens[typIndex + 1] : 'unknown'
  const type = normalizeCandidateType(typeToken)

  if (!ip || ip.includes('.local')) {
    return null
  }

  return {
    ip,
    type,
    protocol,
    raw: candidate,
    isPrivate: isPrivateIp(ip),
  }
}

function normalizeCandidateType(value: string): IceCandidateType {
  switch (value) {
    case 'host':
    case 'srflx':
    case 'relay':
    case 'prflx':
      return value
    default:
      return 'unknown'
  }
}

export async function gatherWebRtcCandidates(
  signal?: AbortSignal,
): Promise<WebRtcGatherResult> {
  if (!isWebRtcSupported()) {
    throw new Error('WebRTC is not supported in this browser')
  }

  const peer = new RTCPeerConnection({
    iceServers: [{ urls: STUN_URLS }],
  })

  const seen = new Set<string>()
  const candidates: GatheredCandidate[] = []

  const cleanup = () => {
    peer.onicecandidate = null
    peer.onicegatheringstatechange = null
    peer.close()
  }

  if (signal?.aborted) {
    cleanup()
    throw new DOMException('Aborted', 'AbortError')
  }

  const abortHandler = () => {
    cleanup()
  }
  signal?.addEventListener('abort', abortHandler, { once: true })

  try {
    peer.createDataChannel('asip-leak-check')

    await new Promise<void>((resolve, reject) => {
      const finish = () => {
        resolve()
      }

      peer.onicecandidate = (event) => {
        if (!event.candidate?.candidate) {
          return
        }
        const parsed = parseCandidate(event.candidate.candidate)
        if (!parsed) {
          return
        }
        const key = `${parsed.type}|${parsed.ip}|${parsed.protocol}`
        if (seen.has(key)) {
          return
        }
        seen.add(key)
        candidates.push(parsed)
      }

      peer.onicegatheringstatechange = () => {
        if (peer.iceGatheringState === 'complete') {
          finish()
        }
      }

      peer
        .createOffer()
        .then((offer) => peer.setLocalDescription(offer))
        .catch(reject)

      // Safety timeout if gathering never completes.
      window.setTimeout(finish, 5000)
    })

    if (signal?.aborted) {
      throw new DOMException('Aborted', 'AbortError')
    }

    candidates.sort((left, right) => {
      if (left.type !== right.type) {
        return left.type.localeCompare(right.type)
      }
      return left.ip.localeCompare(right.ip)
    })

    return { candidates }
  } finally {
    signal?.removeEventListener('abort', abortHandler)
    cleanup()
  }
}
