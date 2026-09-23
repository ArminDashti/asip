/** Shared display helpers for values the API may not know (e.g. an IP that has
 * no AS/country attribution). Missing values render as "Not available" instead
 * of an empty or misleading "0". */

export const NOT_AVAILABLE = 'Not available'

export function isKnownAsn(asn: number): boolean {
  return Number.isFinite(asn) && asn > 0
}

export function isKnownText(value: string | null | undefined): boolean {
  return (value ?? '').trim() !== ''
}

export function formatAsn(asn: number): string {
  return isKnownAsn(asn) ? String(asn) : NOT_AVAILABLE
}

export function formatText(value: string | null | undefined): string {
  return isKnownText(value) ? (value as string).trim() : NOT_AVAILABLE
}

/** CSS class marking a placeholder value, so it can be styled as muted. */
export function emptyClass(known: boolean): string | undefined {
  return known ? undefined : 'is-empty'
}
