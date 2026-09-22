import { TriangleAlert } from "lucide-react"

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { isOverAllocated } from "@/lib/capacity-band"

export function OverAllocationBanner({
  overAllocated,
  warnings = [],
}: {
  overAllocated: boolean
  warnings?: readonly string[]
}) {
  if (!isOverAllocated(overAllocated, warnings)) return null

  return (
    <Alert>
      <TriangleAlert />
      <AlertTitle>Saved with over-allocation</AlertTitle>
      <AlertDescription>
        This allocation was persisted. The person is over 100% on overlapping days. Adjust other
        allocations if you need to bring utilization down.
      </AlertDescription>
    </Alert>
  )
}
