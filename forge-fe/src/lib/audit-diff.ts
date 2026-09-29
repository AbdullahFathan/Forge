import { priorityMap, projectStatusMap, taskStatusMap } from "@/components/common/status-map"

export type FieldChange = {
  key: string
  label: string
  before: string
  after: string
}

const ACTION_LABELS: Record<string, string> = {
  CREATED: "Created",
  UPDATED: "Updated",
  DELETED: "Deleted",
  STATUS_CHANGED: "Status changed",
  TASK_STATUS_CHANGED: "Status changed",
}

export function formatAuditAction(action: string) {
  const known = ACTION_LABELS[action]
  if (known) return known
  const words = action.toLowerCase().split("_").filter(Boolean)
  if (words.length === 0) return action
  return words.map((word, index) => (index === 0 ? word.charAt(0).toUpperCase() + word.slice(1) : word)).join(" ")
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value)
}

function parseRecord(value: unknown): Record<string, unknown> | null {
  if (typeof value === "string") {
    const trimmed = value.trim()
    if (!trimmed) return null
    try {
      return parseRecord(JSON.parse(trimmed))
    } catch {
      return null
    }
  }
  return isRecord(value) ? value : null
}

const HIDDEN_KEY = /password|secret|token|hash/i
const NOISE_KEY = /^(createdAt|updatedAt|deletedAt)$/i

function summarizeRecord(value: Record<string, unknown>) {
  const name = value.Name ?? value.name ?? value.Email ?? value.email ?? value.Code ?? value.code
  if (typeof name === "string" || typeof name === "number") return String(name)
  const id = value.ID ?? value.Id ?? value.id
  if (typeof id === "string" || typeof id === "number") return String(id)
  return `${Object.keys(value).length} fields`
}

function collect(value: unknown, prefix: string, depth: number, out: Map<string, unknown>) {
  if (!isRecord(value) || depth > 0) {
    if (prefix) out.set(prefix, isRecord(value) ? summarizeRecord(value) : value)
    return
  }
  const entries = Object.entries(value).filter(([key]) => !HIDDEN_KEY.test(key) && !NOISE_KEY.test(key))
  if (entries.length === 0) {
    if (prefix) out.set(prefix, value)
    return
  }
  for (const [key, child] of entries) {
    const path = prefix ? `${prefix}.${key}` : key
    if (isRecord(child)) out.set(path, summarizeRecord(child))
    else out.set(path, child)
  }
}

function snapshot(value: unknown) {
  const record = parseRecord(value)
  const fields = new Map<string, unknown>()
  if (record) collect(record, "", 0, fields)
  return fields
}

function same(left: unknown, right: unknown) {
  return JSON.stringify(left) === JSON.stringify(right)
}

export function fieldLabel(path: string) {
  return path
    .split(".")
    .map((part) =>
      part
        .replace(/_/g, " ")
        .replace(/([a-z0-9])([A-Z])/g, "$1 $2")
        .replace(/\bId\b/gi, "ID"),
    )
    .join(" · ")
}

export function redactSecrets(value: unknown): unknown {
  if (Array.isArray(value)) return value.map((item) => redactSecrets(item))
  if (!isRecord(value)) return value
  return Object.fromEntries(
    Object.entries(value)
      .filter(([key]) => !HIDDEN_KEY.test(key))
      .map(([key, child]) => [key, redactSecrets(child)]),
  )
}

export function formatAuditValue(value: unknown) {
  if (value === undefined || value === null || value === "") return "—"
  if (typeof value === "boolean") return value ? "Yes" : "No"
  if (typeof value === "number") return String(value)
  if (typeof value === "string") {
    if (/^\d{4}-\d{2}-\d{2}T/.test(value)) {
      const date = new Date(value)
      if (!Number.isNaN(date.getTime())) return date.toLocaleString()
    }
    const known =
      projectStatusMap[value]?.label ?? taskStatusMap[value]?.label ?? priorityMap[value]?.label
    if (known) return known
    return value.length > 160 ? `${value.slice(0, 157)}…` : value
  }
  if (Array.isArray(value)) {
    if (value.every((item) => typeof item === "string" || typeof item === "number")) {
      const joined = value.join(", ")
      if (!joined) return "—"
      return joined.length > 160 ? `${joined.slice(0, 157)}…` : joined
    }
    return `${value.length} items`
  }
  try {
    const text = JSON.stringify(value)
    return text.length > 160 ? `${text.slice(0, 157)}…` : text
  } catch {
    return String(value)
  }
}

export function fieldChanges(before: unknown, after: unknown): FieldChange[] {
  const previous = snapshot(before)
  const next = snapshot(after)
  const keys = [...new Set([...previous.keys(), ...next.keys()])].sort()
  return keys.flatMap((key) => {
    const left = previous.get(key)
    const right = next.get(key)
    if (same(left, right)) return []
    const beforeText = formatAuditValue(left)
    const afterText = formatAuditValue(right)
    if (beforeText === "—" && afterText === "—") return []
    if (!previous.has(key) && afterText === "—") return []
    if (!next.has(key) && beforeText === "—") return []
    return [{ key, label: fieldLabel(key), before: beforeText, after: afterText }]
  }).filter((change, _, list) => {
    const parent = change.key.match(/^(.*)Id$/i)?.[1]
    return !parent || !list.some((item) => item.key === parent)
  })
}
