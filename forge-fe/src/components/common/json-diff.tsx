import { useState } from "react"

import { Button } from "@/components/ui/button"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { fieldChanges, redactSecrets } from "@/lib/audit-diff"

function pretty(value: unknown) {
  if (value === undefined || value === null || value === "") return "—"
  const source = typeof value === "string" ? (() => {
    try {
      return JSON.parse(value) as unknown
    } catch {
      return value
    }
  })() : value
  try {
    return JSON.stringify(redactSecrets(source), null, 2)
  } catch {
    return String(value)
  }
}

export function JsonDiff({ before, after }: { before?: unknown; after?: unknown }) {
  const [showRecord, setShowRecord] = useState(false)
  const changes = fieldChanges(before, after)
  const hasBefore = changes.some((change) => change.before !== "—")
  const hasAfter = changes.some((change) => change.after !== "—")

  return (
    <div className="flex flex-col gap-3">
      {changes.length === 0 ? (
        <p className="text-sm text-muted-foreground">No field changes in this record.</p>
      ) : (
        <Table className="table-fixed">
          <TableHeader>
            <TableRow>
              <TableHead className="w-1/3 px-3 py-2 whitespace-normal!">Field</TableHead>
              {hasBefore ? <TableHead className="px-3 py-2 whitespace-normal!">Before</TableHead> : null}
              {hasAfter ? <TableHead className="px-3 py-2 whitespace-normal!">After</TableHead> : null}
            </TableRow>
          </TableHeader>
          <TableBody>
            {changes.map((change) => (
              <TableRow key={change.key}>
                <TableCell className="px-3 py-2 font-medium whitespace-normal!">{change.label}</TableCell>
                {hasBefore ? (
                  <TableCell className="px-3 py-2 break-words whitespace-normal! text-muted-foreground">
                    {change.before}
                  </TableCell>
                ) : null}
                {hasAfter ? (
                  <TableCell className="px-3 py-2 break-words whitespace-normal!">{change.after}</TableCell>
                ) : null}
              </TableRow>
            ))}
          </TableBody>
        </Table>
      )}
      <div>
        <Button type="button" variant="ghost" size="sm" onClick={() => setShowRecord((open) => !open)}>
          {showRecord ? "Hide full record" : "Full record"}
        </Button>
        {showRecord ? (
          <div className="mt-2 grid gap-3 md:grid-cols-2">
            <div className="flex flex-col gap-1">
              <p className="text-xs font-medium text-muted-foreground">Before</p>
              <pre className="max-h-64 overflow-auto rounded-md bg-muted p-2 font-mono text-xs">
                {pretty(before)}
              </pre>
            </div>
            <div className="flex flex-col gap-1">
              <p className="text-xs font-medium text-muted-foreground">After</p>
              <pre className="max-h-64 overflow-auto rounded-md bg-muted p-2 font-mono text-xs">
                {pretty(after)}
              </pre>
            </div>
          </div>
        ) : null}
      </div>
    </div>
  )
}
