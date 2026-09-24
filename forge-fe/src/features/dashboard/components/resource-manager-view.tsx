import { TriangleAlert } from "lucide-react"
import { useMemo } from "react"
import { useQuery } from "@tanstack/react-query"

import { CapacityCell } from "@/components/common/capacity-cell"
import { EmptyState } from "@/components/common/empty-state"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { getAvailability, getOverloadAlerts } from "@/features/resources/api/capacity"
import { defaultCapacityRange } from "@/lib/date-range"
import { queryKeys } from "@/services/query/query-keys"
import { ApiError } from "@/types/api"

const AVAILABLE_SLICE = 8

function queryError(error: unknown, fallback: string) {
  return error instanceof ApiError ? error.message : error ? fallback : null
}

export function ResourceManagerView() {
  const range = useMemo(
    () => ({ ...defaultCapacityRange(), granularity: "week" as const }),
    [],
  )
  const alerts = useQuery({
    queryKey: queryKeys.overloadAlerts,
    queryFn: getOverloadAlerts,
  })
  const availability = useQuery({
    queryKey: queryKeys.availability(range),
    queryFn: () => getAvailability(range),
  })

  const alertsError = queryError(alerts.error, "Could not load overload alerts")
  const availabilityError = queryError(availability.error, "Could not load availability")
  const availableRows = (availability.data ?? []).slice(0, AVAILABLE_SLICE)
  const alertRows = alerts.data ?? []

  return (
    <div className="grid gap-4 lg:grid-cols-2">
      <Card>
        <CardHeader>
          <CardTitle>Overload (14 days)</CardTitle>
          <CardDescription>
            People whose planned load exceeds 100% on any day in the next two weeks.
          </CardDescription>
        </CardHeader>
        <CardContent>
          {alertsError ? (
            <Alert variant="destructive">
              <AlertTitle>Could not load overload alerts</AlertTitle>
              <AlertDescription>{alertsError}</AlertDescription>
            </Alert>
          ) : null}
          {alerts.isLoading ? (
            <div className="flex flex-col gap-2">
              {Array.from({ length: 3 }, (_, index) => (
                <Skeleton key={index} className="h-8 w-full" />
              ))}
            </div>
          ) : null}
          {!alerts.isLoading && !alertsError && alertRows.length === 0 ? (
            <EmptyState
              title="No overload alerts"
              description="Nobody is over 100% in the next two weeks."
            />
          ) : null}
          {!alerts.isLoading && !alertsError && alertRows.length > 0 ? (
            <ul className="flex flex-col gap-2">
              {alertRows.map((row) => (
                <li key={row.userId} className="flex items-center justify-between gap-2 text-sm">
                  <span>{row.name}</span>
                  <span className="inline-flex items-center gap-1 tabular-nums text-cap-over-foreground">
                    <TriangleAlert className="size-3.5" aria-hidden />
                    <span>
                      {Math.round(row.maxPercent)}% over allocated
                    </span>
                  </span>
                </li>
              ))}
            </ul>
          ) : null}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Available</CardTitle>
          <CardDescription>
            People under 80% utilization in the next 4 weeks (highest remaining first).
          </CardDescription>
        </CardHeader>
        <CardContent>
          {availabilityError ? (
            <Alert variant="destructive">
              <AlertTitle>Could not load availability</AlertTitle>
              <AlertDescription>{availabilityError}</AlertDescription>
            </Alert>
          ) : null}
          {availability.isLoading ? (
            <div className="flex flex-col gap-2">
              {Array.from({ length: 4 }, (_, index) => (
                <Skeleton key={index} className="h-8 w-full" />
              ))}
            </div>
          ) : null}
          {!availability.isLoading && !availabilityError && availableRows.length === 0 ? (
            <EmptyState
              title="No available people"
              description="Everyone in this range is at 80% or more."
            />
          ) : null}
          {!availability.isLoading && !availabilityError && availableRows.length > 0 ? (
            <ul className="flex flex-col gap-2">
              {availableRows.map((row) => (
                <li key={row.userId} className="flex items-center justify-between gap-2 text-sm">
                  <span>{row.name}</span>
                  <CapacityCell percent={row.utilizationPercent} band={row.band} />
                </li>
              ))}
            </ul>
          ) : null}
        </CardContent>
      </Card>
    </div>
  )
}
