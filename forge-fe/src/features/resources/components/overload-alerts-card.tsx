import { useQuery } from "@tanstack/react-query"

import { EmptyState } from "@/components/common/empty-state"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { getOverloadAlerts } from "@/features/resources/api/capacity"
import { queryKeys } from "@/services/query/query-keys"

export function OverloadAlertsCard() {
  const alerts = useQuery({
    queryKey: queryKeys.overloadAlerts,
    queryFn: getOverloadAlerts,
  })

  const rows = alerts.data ?? []

  return (
    <Card>
      <CardHeader>
        <CardTitle>Overload in the next 2 weeks</CardTitle>
        <CardDescription>People whose planned load exceeds 100% on any day.</CardDescription>
      </CardHeader>
      <CardContent>
        {alerts.isLoading ? (
          <p className="text-sm text-muted-foreground">Loading alerts…</p>
        ) : rows.length === 0 ? (
          <EmptyState title="No overload alerts" description="Nobody is over 100% in the next two weeks." />
        ) : (
          <ul className="flex flex-col gap-2">
            {rows.map((row) => (
              <li key={row.userId} className="flex items-center justify-between gap-2 text-sm">
                <span>{row.name}</span>
                <span className="tabular-nums text-cap-over-foreground">{Math.round(row.maxPercent)}%</span>
              </li>
            ))}
          </ul>
        )}
      </CardContent>
    </Card>
  )
}
