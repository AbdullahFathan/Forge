import { Badge } from "@/components/ui/badge"
import { priorityMap } from "@/components/common/status-map"

export function PriorityBadge({ value }: { value: string }) {
  const visual = priorityMap[value]
  if (!visual) {
    return <Badge variant="secondary">{value}</Badge>
  }
  const Icon = visual.icon
  return (
    <Badge variant={visual.variant} className={visual.className}>
      <Icon data-icon="inline-start" />
      {visual.label}
    </Badge>
  )
}
