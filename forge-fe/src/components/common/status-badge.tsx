import { Badge } from "@/components/ui/badge"
import { projectStatusMap, taskStatusMap, type StatusVisual } from "@/components/common/status-map"

const maps = { project: projectStatusMap, task: taskStatusMap }

export function StatusBadge({
  kind,
  value,
}: {
  kind: keyof typeof maps
  value: string
}) {
  const visual: StatusVisual | undefined = maps[kind][value]
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
