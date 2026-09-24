import { ClockAlert } from "lucide-react"

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import {
  calendarProgress,
  completionRatio,
  isProjectOverdue,
  isValidIsoDate,
} from "@/features/projects/lib/project-timeline"
import { toIsoDate } from "@/lib/date-range"

export function ProjectOverviewTimeline({
  startDate,
  targetEndDate,
  completionPercent,
  status,
  todayIso = toIsoDate(new Date()),
}: {
  startDate: string
  targetEndDate: string
  completionPercent: number
  status: string
  todayIso?: string
}) {
  const datesOk = isValidIsoDate(startDate) && isValidIsoDate(targetEndDate)
  const overdue = isProjectOverdue({ targetEndDate, status, todayIso })
  const todayPos = datesOk ? calendarProgress(startDate, targetEndDate, todayIso) : null
  const done = completionRatio(completionPercent)
  const percentLabel = Math.round(Number.isFinite(completionPercent) ? completionPercent : 0)

  return (
    <Card>
      <CardHeader className="flex flex-row items-start justify-between gap-2">
        <CardTitle>Timeline</CardTitle>
        {overdue ? (
          <p className="flex items-center gap-1 text-sm text-destructive">
            <ClockAlert className="size-4" aria-hidden />
            Overdue
          </p>
        ) : null}
      </CardHeader>
      <CardContent className="flex flex-col gap-3">
        {!datesOk ? (
          <p className="text-sm text-muted-foreground">Dates unavailable</p>
        ) : (
          <>
            <p className="sr-only">
              {`Start ${startDate}. Today ${todayIso}. Target end ${targetEndDate}. Completion ${percentLabel} percent.`}
              {overdue ? " Overdue." : ""}
            </p>
            <div className="relative h-3 rounded-full bg-muted">
              <div
                role="progressbar"
                aria-valuemin={0}
                aria-valuemax={100}
                aria-valuenow={percentLabel}
                aria-label="Project completion"
                className="absolute inset-y-0 left-0 rounded-full bg-primary"
                style={{ width: `${done * 100}%` }}
              />
              {todayPos != null ? (
                <div
                  className="absolute top-1/2 z-10 h-5 w-px -translate-x-1/2 -translate-y-1/2 bg-foreground"
                  style={{ left: `${todayPos * 100}%` }}
                  aria-hidden
                />
              ) : null}
            </div>
            {overdue ? <div className="h-1 rounded-full bg-danger-subtle" aria-hidden /> : null}
            <div className="grid grid-cols-3 gap-2 text-xs">
              <p>
                <span className="block text-muted-foreground">Start</span>
                <span className="font-mono tabular-nums">{startDate}</span>
              </p>
              <p className="text-center">
                <span className="block text-muted-foreground">Today</span>
                <span className="font-mono tabular-nums">{todayIso}</span>
              </p>
              <p className="text-right">
                <span className="block text-muted-foreground">Target end</span>
                <span className="font-mono tabular-nums">{targetEndDate}</span>
              </p>
            </div>
            <p className="text-sm text-muted-foreground">
              Progress <span className="font-mono tabular-nums">{percentLabel}%</span>
            </p>
          </>
        )}
      </CardContent>
    </Card>
  )
}
