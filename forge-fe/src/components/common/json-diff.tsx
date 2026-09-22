function pretty(value: unknown) {
  if (value === undefined || value === null || value === "") return "—"
  if (typeof value === "string") {
    try {
      return JSON.stringify(JSON.parse(value), null, 2)
    } catch {
      return value
    }
  }
  try {
    return JSON.stringify(value, null, 2)
  } catch {
    return String(value)
  }
}

export function JsonDiff({ before, after }: { before?: unknown; after?: unknown }) {
  return (
    <div className="grid gap-3 md:grid-cols-2">
      <div className="flex flex-col gap-1">
        <p className="text-xs font-medium text-muted-foreground">Before</p>
        <pre className="overflow-auto rounded-md bg-muted p-2 font-mono text-xs">{pretty(before)}</pre>
      </div>
      <div className="flex flex-col gap-1">
        <p className="text-xs font-medium text-muted-foreground">After</p>
        <pre className="overflow-auto rounded-md bg-muted p-2 font-mono text-xs">{pretty(after)}</pre>
      </div>
    </div>
  )
}
