import { Gauge, Leaf, TriangleAlert } from "lucide-react"

import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip"
import { bandClassName, bandLabel } from "@/lib/capacity-band"
import { cn } from "@/lib/utils"
import type { AllocationBreakdown } from "@/types/resource"

export function CapacityCell({
  percent,
  band,
  breakdown,
}: {
  percent: number
  band: string
  breakdown?: AllocationBreakdown[]
}) {
  const label = `${Math.round(percent)}% · ${bandLabel(band)}`
  const cell = (
    <span
      className={cn(
        "inline-flex min-w-14 justify-center rounded-md px-1.5 py-1 font-medium tabular-nums",
        bandClassName(band),
      )}
    >
      {Math.round(percent)}%
    </span>
  )

  return (
    <Tooltip>
      <TooltipTrigger>{cell}</TooltipTrigger>
      <TooltipContent>
        <div className="flex flex-col gap-1">
          <span>{label}</span>
          {breakdown?.map((item) => (
            <span key={item.projectId}>
              {item.projectName ?? item.projectId}: {Math.round(item.allocationPercent)}%
            </span>
          ))}
        </div>
      </TooltipContent>
    </Tooltip>
  )
}

export function CapacityLegend() {
  return (
    <div className="flex flex-wrap items-center gap-3 text-xs text-muted-foreground">
      <span className="inline-flex items-center gap-1">
        <Leaf className="text-cap-available-foreground" />
        Available &lt; 80%
      </span>
      <span className="inline-flex items-center gap-1">
        <Gauge className="text-cap-full-foreground" />
        Full 80–100%
      </span>
      <span className="inline-flex items-center gap-1">
        <TriangleAlert className="text-cap-over-foreground" />
        Over &gt; 100%
      </span>
    </div>
  )
}
