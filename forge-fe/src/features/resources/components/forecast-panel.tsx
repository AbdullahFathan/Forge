import { useQuery } from "@tanstack/react-query"

import { CapacityCell } from "@/components/common/capacity-cell"
import { EmptyState } from "@/components/common/empty-state"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { getCapacity } from "@/features/resources/api/capacity"
import { OverloadAlertsCard } from "@/features/resources/components/overload-alerts-card"
import { queryKeys } from "@/services/query/query-keys"
import { ApiError } from "@/types/api"
import type { CapacityQuery } from "@/features/resources/api/capacity"

export function ForecastPanel({ params }: { params: CapacityQuery }) {
  const forecast = useQuery({
    queryKey: queryKeys.capacity(params),
    queryFn: () => getCapacity(params),
  })

  if (forecast.isLoading) {
    return (
      <div className="grid gap-2">
        <Skeleton className="h-24 w-full" />
        <Skeleton className="h-24 w-full" />
      </div>
    )
  }

  if (forecast.error) {
    const message = forecast.error instanceof ApiError ? forecast.error.message : "Could not load forecast"
    return (
      <Alert variant="destructive">
        <AlertTitle>Could not load forecast</AlertTitle>
        <AlertDescription>{message}</AlertDescription>
      </Alert>
    )
  }

  const data = forecast.data
  if (!data?.users.length) {
    return <EmptyState title="No capacity data" description="No active people match this range." />
  }

  return (
    <div className="flex flex-col gap-3">
      <OverloadAlertsCard />
      {data.users.map((user) => (
        <Card key={user.userId}>
          <CardHeader>
            <CardTitle>{user.name}</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex gap-2 overflow-x-auto">
              {user.buckets.map((bucket) => (
                <div key={bucket.periodKey} className="flex min-w-20 flex-col items-center gap-1">
                  <span className="text-xs text-muted-foreground">{bucket.periodKey}</span>
                  <CapacityCell
                    percent={bucket.utilizationPercent}
                    band={bucket.band}
                    breakdown={bucket.allocationBreakdown}
                  />
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      ))}
    </div>
  )
}
