import { useEffect, useState, type ReactNode } from 'react'
import {
  applyTheme,
  cycleThemePreference,
  resolveEffectiveTheme,
  resolveThemePreference,
  subscribeSystemTheme,
  type ThemePreference,
} from '../lib/theme'

const BASE = (import.meta.env.BASE_URL || '/').replace(/\/$/, '')

const NAV_ITEMS = [
  { href: '/', label: 'Your IP' },
  { href: '/ip', label: 'IP' },
  { href: '/whois', label: 'Whois' },
  { href: '/nslookup', label: 'NsLookup' },
  { href: '/dns-leak', label: 'DNS Leak' },
  { href: '/webrtc', label: 'WebRTC Leak' },
  { href: '/headers', label: 'Headers' },
  { href: '/about', label: 'About me' },
] as const

function withBase(href: string): string {
  if (!BASE || BASE === '/') return href
  if (href === '/') return `${BASE}/`
  return `${BASE}${href}`
}

function appPath(pathname: string): string {
  const normalized = pathname.replace(/\/+$/, '') || '/'
  if (BASE && normalized.startsWith(BASE)) {
    return normalized.slice(BASE.length) || '/'
  }
  return normalized
}

const PREFERENCE_LABELS: Record<ThemePreference, string> = {
  system: 'System',
  light: 'Light',
  dark: 'Dark',
}

function nextPreferenceLabel(preference: ThemePreference): string {
  if (preference === 'system') {
    return 'Light'
  }
  if (preference === 'light') {
    return 'Dark'
  }
  return 'System'
}

function GitHubIcon() {
  return (
    <svg
      className="site-footer-github-icon"
      viewBox="0 0 16 16"
      width="22"
      height="22"
      aria-hidden="true"
      focusable="false"
    >
      <path
        fill="currentColor"
        d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27s1.36.09 2 .27c1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.01 8.01 0 0 0 16 8c0-4.42-3.58-8-8-8z"
      />
    </svg>
  )
}

export function AppShell({ children }: { children: ReactNode }) {
  const pathname = appPath(window.location.pathname)
  const [preference, setPreference] = useState<ThemePreference>(() =>
    resolveThemePreference(),
  )

  useEffect(() => {
    applyTheme(resolveEffectiveTheme(preference))
    if (preference !== 'system') {
      return
    }
    return subscribeSystemTheme((theme) => {
      applyTheme(theme)
    })
  }, [preference])

  function handleThemeCycle() {
    setPreference((current) => cycleThemePreference(current))
  }

  const preferenceLabel = PREFERENCE_LABELS[preference]
  const nextLabel = nextPreferenceLabel(preference)

  return (
    <div className="app-shell">
      <header className="site-header">
        <nav className="site-nav" aria-label="Primary">
          {NAV_ITEMS.map((item) => {
            const isActive = pathname === item.href
            return (
              <a
                key={item.href}
                className={isActive ? 'site-nav-link is-active' : 'site-nav-link'}
                href={withBase(item.href)}
                aria-current={isActive ? 'page' : undefined}
              >
                {item.label}
              </a>
            )
          })}
        </nav>
        <div className="site-header-actions">
          <button
            type="button"
            className="theme-toggle"
            onClick={handleThemeCycle}
            aria-label={`Theme preference: ${preferenceLabel}. Switch to ${nextLabel.toLowerCase()}`}
            title={`Theme: ${preferenceLabel} (next: ${nextLabel})`}
          >
            {preferenceLabel}
          </button>
          <a className="site-logo" href={withBase('/')}>
            ASIP
          </a>
        </div>
      </header>

      <main className="site-main">{children}</main>

      <footer className="site-footer">
        <a
          className="site-footer-github"
          href="https://github.com/ArminDashti"
          target="_blank"
          rel="noopener noreferrer"
          aria-label="Armin Dashti on GitHub"
        >
          <GitHubIcon />
        </a>
        <p className="site-footer-copy">
          © 2026 Dashti Technologies (Armin Dashti)
        </p>
        <span className="site-footer-spacer" aria-hidden="true" />
      </footer>
    </div>
  )
}
