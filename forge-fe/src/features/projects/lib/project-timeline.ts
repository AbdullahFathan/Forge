const ISO_DATE = /^\d{4}-\d{2}-\d{2}$/

export function isValidIsoDate(value: string | undefined | null): value is string {
  return typeof value === "string" && ISO_DATE.test(value)
}

export function isProjectOverdue({
  targetEndDate,
  status,
  todayIso,
}: {
  targetEndDate: string | undefined | null
  status: string
  todayIso: string
}) {
  if (!isValidIsoDate(targetEndDate) || !isValidIsoDate(todayIso)) return false
  if (status === "COMPLETED" || status === "ARCHIVED") return false
  return targetEndDate < todayIso
}

export function clamp01(value: number) {
  if (!Number.isFinite(value)) return 0
  return Math.min(1, Math.max(0, value))
}

function parseUtcDay(iso: string) {
  if (!isValidIsoDate(iso)) return null
  const t = Date.parse(`${iso}T00:00:00`)
  return Number.isNaN(t) ? null : t
}

/** Calendar position of today between start and end, 0–1. Null if dates are invalid. */
export function calendarProgress(start: string, end: string, today: string) {
  const s = parseUtcDay(start)
  const e = parseUtcDay(end)
  const t = parseUtcDay(today)
  if (s === null || e === null || t === null) return null
  if (e <= s) return t >= e ? 1 : 0
  return clamp01((t - s) / (e - s))
}

export function completionRatio(percent: number) {
  return clamp01(percent / 100)
}
