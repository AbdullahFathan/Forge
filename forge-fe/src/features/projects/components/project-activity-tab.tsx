import { useQuery } from "@tanstack/react-query"
import { useState } from "react"

import { EntityTable } from "@/components/common/entity-table"
import { JsonDiff } from "@/components/common/json-diff"
import { listProjectActivity } from "@/features/projects/api/projects"
import { formatAuditAction } from "@/lib/audit-diff"
import { DEFAULT_PAGE_SIZE } from "@/lib/constants"
import { ApiError } from "@/types/api"
import { queryKeys } from "@/services/query/query-keys"

export function ProjectActivityTab({ projectId }: { projectId: string }) {
  const [page, setPage] = useState(1)
  const list = useQuery({
    queryKey: queryKeys.projectActivity(projectId, { page, pageSize: DEFAULT_PAGE_SIZE }),
    queryFn: () => listProjectActivity(projectId, page, DEFAULT_PAGE_SIZE),
  })

  const error =
    list.error instanceof ApiError ? list.error.message : list.error ? "Could not load activity" : null

  return (
    <EntityTable
      columns={[
        {
          key: "when",
          header: "When",
          className: "px-3 py-2 font-mono tabular-nums whitespace-nowrap",
          cell: (row) => (
            <time dateTime={row.createdAt}>{new Date(row.createdAt).toLocaleString()}</time>
          ),
        },
        { key: "who", header: "Who", cell: (row) => row.userName || "—" },
        { key: "action", header: "Action", cell: (row) => formatAuditAction(row.action) },
        {
          key: "entity",
          header: "Entity",
          cell: (row) => (
            <span className="flex min-w-0 flex-col">
              <span className="truncate">{row.entityName || row.entityType}</span>
              {row.entityName && row.entityType ? (
                <span className="text-xs text-muted-foreground">{row.entityType}</span>
              ) : null}
            </span>
          ),
        },
      ]}
      rows={list.data?.items ?? []}
      rowKey={(row) => row.id}
      isLoading={list.isLoading}
      error={error}
      emptyTitle="No activity yet"
      emptyDescription="Writes on this project and its tasks appear here."
      page={page}
      pageSize={list.data?.pageSize ?? DEFAULT_PAGE_SIZE}
      totalItems={list.data?.totalItems ?? 0}
      onPageChange={setPage}
      renderExpanded={(row) => <JsonDiff before={row.before} after={row.after} />}
    />
  )
}
