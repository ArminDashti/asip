export type Theme = 'dark' | 'light'
export type ThemePreference = 'system' | Theme

export const THEME_STORAGE_KEY = 'asip-theme'

export function isTheme(value: string | null | undefined): value is Theme {
  return value === 'dark' || value === 'light'
}

export function isThemePreference(
  value: string | null | undefined,
): value is ThemePreference {
  return value === 'system' || isTheme(value)
}

export function readStoredThemePreference(): ThemePreference | null {
  try {
    const stored = localStorage.getItem(THEME_STORAGE_KEY)
    return isThemePreference(stored) ? stored : null
  } catch {
    return null
  }
}

export function getSystemTheme(): Theme {
  try {
    return window.matchMedia('(prefers-color-scheme: dark)').matches
      ? 'dark'
      : 'light'
  } catch {
    return 'dark'
  }
}

export function resolveThemePreference(): ThemePreference {
  return readStoredThemePreference() ?? 'system'
}

export function resolveEffectiveTheme(
  preference: ThemePreference = resolveThemePreference(),
): Theme {
  return preference === 'system' ? getSystemTheme() : preference
}

export function applyTheme(theme: Theme): void {
  document.documentElement.dataset.theme = theme
}

export function persistThemePreference(preference: ThemePreference): void {
  try {
    localStorage.setItem(THEME_STORAGE_KEY, preference)
  } catch {
    // ignore quota / private mode failures
  }
}

export function setThemePreference(preference: ThemePreference): void {
  applyTheme(resolveEffectiveTheme(preference))
  persistThemePreference(preference)
}

export function cycleThemePreference(
  current: ThemePreference,
): ThemePreference {
  const next: ThemePreference =
    current === 'system' ? 'light' : current === 'light' ? 'dark' : 'system'
  setThemePreference(next)
  return next
}

export function subscribeSystemTheme(onChange: (theme: Theme) => void): () => void {
  let media: MediaQueryList
  try {
    media = window.matchMedia('(prefers-color-scheme: dark)')
  } catch {
    return () => {}
  }

  const handler = (event: MediaQueryListEvent) => {
    onChange(event.matches ? 'dark' : 'light')
  }

  if (typeof media.addEventListener === 'function') {
    media.addEventListener('change', handler)
    return () => media.removeEventListener('change', handler)
  }

  media.addListener(handler)
  return () => media.removeListener(handler)
}
