function pad(value: number) {
  return String(value).padStart(2, "0")
}

export function toIsoDate(date: Date) {
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}`
}

export function addDays(iso: string, days: number) {
  const date = new Date(`${iso}T00:00:00`)
  date.setDate(date.getDate() + days)
  return toIsoDate(date)
}

/** Inclusive 28-day window starting today (4 weeks). */
export function defaultCapacityRange() {
  const from = toIsoDate(new Date())
  return { from, to: addDays(from, 27) }
}
