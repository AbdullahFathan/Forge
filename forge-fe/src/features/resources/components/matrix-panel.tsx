import { useQuery } from "@tanstack/react-query"

import { CapacityCell } from "@/components/common/capacity-cell"
import { EmptyState } from "@/components/common/empty-state"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { Skeleton } from "@/components/ui/skeleton"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { getMatrix } from "@/features/resources/api/capacity"
import type { CapacityQuery } from "@/features/resources/api/capacity"
import { queryKeys } from "@/services/query/query-keys"
import { ApiError } from "@/types/api"

export function MatrixPanel({ params }: { params: CapacityQuery }) {
  const matrix = useQuery({
    queryKey: queryKeys.matrix(params),
    queryFn: () => getMatrix(params),
  })

  if (matrix.isLoading) {
    return <Skeleton className="h-64 w-full" />
  }

  if (matrix.error) {
    const err = matrix.error
    const needsDepartment =
      err instanceof ApiError &&
      err.code === "VALIDATION_ERROR" &&
      err.message.toLowerCase().includes("department")
    return (
      <Alert variant="destructive">
        <AlertTitle>{needsDepartment ? "Department required" : "Could not load matrix"}</AlertTitle>
        <AlertDescription>
          {err instanceof ApiError
            ? err.message
            : "Could not load the allocation matrix."}
        </AlertDescription>
      </Alert>
    )
  }

  const data = matrix.data
  if (!data?.users.length) {
    return <EmptyState title="Empty matrix" description="No people match this range." />
  }

  return (
    <div className="overflow-x-auto">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead className="sticky left-0 z-10 min-w-40 bg-background px-3 py-2">Person</TableHead>
            {data.periods.map((period) => (
              <TableHead key={period.key} className="min-w-24 px-3 py-2">
                {period.key}
              </TableHead>
            ))}
          </TableRow>
        </TableHeader>
        <TableBody>
          {data.users.map((user) => (
            <TableRow key={user.id}>
              <TableCell className="sticky left-0 z-10 bg-background px-3 py-2">{user.name}</TableCell>
              {data.periods.map((period) => {
                const cell = data.cells[user.id]?.[period.key]
                return (
                  <TableCell key={period.key} className="px-3 py-2">
                    {cell ? (
                      <CapacityCell percent={cell.utilizationPercent} band={cell.band} />
                    ) : (
                      "—"
                    )}
                  </TableCell>
                )
              })}
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}
