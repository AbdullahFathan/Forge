import { useQuery } from "@tanstack/react-query"

import { CapacityCell } from "@/components/common/capacity-cell"
import { EntityTable } from "@/components/common/entity-table"
import { Button } from "@/components/ui/button"
import { getAvailability } from "@/features/resources/api/capacity"
import type { CapacityQuery } from "@/features/resources/api/capacity"
import { queryKeys } from "@/services/query/query-keys"
import { ApiError } from "@/types/api"
import type { AvailabilityItem } from "@/types/resource"

export function AvailabilityPanel({
  params,
  canAllocate,
  onAllocate,
}: {
  params: CapacityQuery
  canAllocate: boolean
  onAllocate: (row: AvailabilityItem) => void
}) {
  const list = useQuery({
    queryKey: queryKeys.availability(params),
    queryFn: () => getAvailability(params),
  })

  const rows = list.data ?? []

  return (
    <EntityTable
      columns={[
        { key: "name", header: "Person", cell: (row) => row.name },
        {
          key: "util",
          header: "Utilization",
          cell: (row) => <CapacityCell percent={row.utilizationPercent} band={row.band} />,
        },
        {
          key: "remain",
          header: "Remaining",
          cell: (row) => <span className="tabular-nums">{Math.round(row.remainingPercent)}%</span>,
        },
        {
          key: "actions",
          header: "",
          cell: (row) =>
            canAllocate ? (
              <Button size="sm" variant="outline" onClick={() => onAllocate(row)}>
                Allocate
              </Button>
            ) : null,
        },
      ]}
      rows={rows}
      rowKey={(row) => row.userId}
      isLoading={list.isLoading}
      error={list.error instanceof ApiError ? list.error.message : list.error ? "Could not load availability" : null}
      emptyTitle="No available people"
      emptyDescription="Everyone in this range is at 80% or more, or no one matches the filters."
      page={1}
      pageSize={Math.max(rows.length, 1)}
      totalItems={rows.length}
      onPageChange={() => undefined}
    />
  )
}
