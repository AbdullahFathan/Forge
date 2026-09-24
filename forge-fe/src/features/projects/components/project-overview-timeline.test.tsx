import { describe, expect, it } from "vitest"

import { ProjectOverviewTimeline } from "@/features/projects/components/project-overview-timeline"

describe("ProjectOverviewTimeline", () => {
  it("shows Overdue when the target date has passed", () => {
    const node = ProjectOverviewTimeline({
      startDate: "2026-01-01",
      targetEndDate: "2026-01-10",
      completionPercent: 40,
      status: "ACTIVE",
      todayIso: "2026-01-20",
    })
    expect(JSON.stringify(node)).toContain("Overdue")
  })

  it("does not crash without dates", () => {
    const node = ProjectOverviewTimeline({
      startDate: "",
      targetEndDate: "",
      completionPercent: 0,
      status: "DRAFT",
      todayIso: "2026-01-20",
    })
    expect(JSON.stringify(node)).toContain("Dates unavailable")
    expect(JSON.stringify(node)).not.toContain("Overdue")
  })
})
