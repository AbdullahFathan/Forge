export function compactParams(params: Record<string, unknown>) {
  const out: Record<string, string | number | boolean> = {}
  for (const [key, value] of Object.entries(params)) {
    if (value === undefined || value === null || value === "" || value === "all") continue
    if (typeof value === "string" || typeof value === "number" || typeof value === "boolean") {
      out[key] = value
    }
  }
  return out
}

export function taskFilterParams(input: {
  page: number
  pageSize: number
  status?: string
  assignee?: string
  priority?: string
  dueFrom?: string
  dueTo?: string
  label?: string
  include?: string
}) {
  return compactParams(input)
}
