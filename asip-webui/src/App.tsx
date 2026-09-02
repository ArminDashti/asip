import { AppShell } from './components/AppShell'
import { AboutPage } from './pages/AboutPage'
import { DnsLeakPage } from './pages/DnsLeakPage'
import { HeadersPage } from './pages/HeadersPage'
import { HomePage } from './pages/HomePage'
import { IpPage } from './pages/IpPage'
import { NslookupPage } from './pages/NslookupPage'
import { WebRtcPage } from './pages/WebRtcPage'
import { WhoisPage } from './pages/WhoisPage'

function appPath(): string {
  const base = (import.meta.env.BASE_URL || '/').replace(/\/$/, '')
  const pathname = window.location.pathname.replace(/\/+$/, '') || '/'
  if (base && pathname.startsWith(base)) {
    return pathname.slice(base.length) || '/'
  }
  return pathname
}

function resolvePage() {
  const pathname = appPath()

  if (pathname === '/ip') {
    return <IpPage />
  }

  if (pathname === '/whois') {
    return <WhoisPage />
  }

  if (pathname === '/nslookup') {
    return <NslookupPage />
  }

  if (pathname === '/dns-leak') {
    return <DnsLeakPage />
  }

  if (pathname === '/webrtc') {
    return <WebRtcPage />
  }

  if (pathname === '/headers') {
    return <HeadersPage />
  }

  if (pathname === '/about') {
    return <AboutPage />
  }

  if (pathname !== '/') {
    const base = (import.meta.env.BASE_URL || '/').replace(/\/$/, '')
    window.location.replace(`${base}/`)
    return null
  }

  return <HomePage />
}

export default function App() {
  const page = resolvePage()
  if (page === null) {
    return null
  }

  return <AppShell>{page}</AppShell>
}
