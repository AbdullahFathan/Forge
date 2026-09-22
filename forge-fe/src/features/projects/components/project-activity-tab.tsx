import { useQuery } from "@tanstack/react-query"
import { useState } from "react"

import { EntityTable } from "@/components/common/entity-table"
import { JsonDiff } from "@/components/common/json-diff"
import { listProjectActivity } from "@/features/projects/api/projects"
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
        { key: "action", header: "Action", cell: (row) => row.action },
        { key: "entity", header: "Entity", cell: (row) => `${row.entityType} · ${row.entityId.slice(0, 8)}` },
        {
          key: "when",
          header: "When",
          className: "px-3 py-2 font-mono tabular-nums",
          cell: (row) => new Date(row.createdAt).toLocaleString(),
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
