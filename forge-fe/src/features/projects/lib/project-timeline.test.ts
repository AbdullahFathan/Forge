import { describe, expect, it } from "vitest"

import {
  calendarProgress,
  clamp01,
  completionRatio,
  isProjectOverdue,
  isValidIsoDate,
} from "@/features/projects/lib/project-timeline"

describe("isProjectOverdue", () => {
  it("flags an active project past target end", () => {
    expect(
      isProjectOverdue({
        targetEndDate: "2026-01-10",
        status: "ACTIVE",
        todayIso: "2026-01-11",
      }),
    ).toBe(true)
  })

  it("does not flag when the target is today", () => {
    expect(
      isProjectOverdue({
        targetEndDate: "2026-01-11",
        status: "ACTIVE",
        todayIso: "2026-01-11",
      }),
    ).toBe(false)
  })

  it("does not flag completed or archived past the deadline", () => {
    expect(
      isProjectOverdue({
        targetEndDate: "2026-01-01",
        status: "COMPLETED",
        todayIso: "2026-01-11",
      }),
    ).toBe(false)
    expect(
      isProjectOverdue({
        targetEndDate: "2026-01-01",
        status: "ARCHIVED",
        todayIso: "2026-01-11",
      }),
    ).toBe(false)
  })

  it("is false for invalid dates", () => {
    expect(
      isProjectOverdue({
        targetEndDate: "",
        status: "ACTIVE",
        todayIso: "2026-01-11",
      }),
    ).toBe(false)
  })
})

describe("calendarProgress", () => {
  it("puts today in the middle of the range", () => {
    expect(calendarProgress("2026-01-01", "2026-01-11", "2026-01-06")).toBeCloseTo(0.5)
  })

  it("clamps before start to 0", () => {
    expect(calendarProgress("2026-02-01", "2026-02-28", "2026-01-15")).toBe(0)
  })

  it("clamps after end to 1", () => {
    expect(calendarProgress("2026-01-01", "2026-01-10", "2026-02-01")).toBe(1)
  })

  it("returns null for invalid dates", () => {
    expect(calendarProgress("not-a-date", "2026-01-10", "2026-01-05")).toBeNull()
    expect(calendarProgress("2026-01-01", "", "2026-01-05")).toBeNull()
  })

  it("handles end on or before start", () => {
    expect(calendarProgress("2026-01-10", "2026-01-10", "2026-01-09")).toBe(0)
    expect(calendarProgress("2026-01-10", "2026-01-10", "2026-01-10")).toBe(1)
  })
})

describe("clamp01 and completionRatio", () => {
  it("clamps out of range values", () => {
    expect(clamp01(-1)).toBe(0)
    expect(clamp01(2)).toBe(1)
    expect(completionRatio(40)).toBeCloseTo(0.4)
    expect(completionRatio(150)).toBe(1)
  })
})

describe("isValidIsoDate", () => {
  it("accepts YYYY-MM-DD", () => {
    expect(isValidIsoDate("2026-09-24")).toBe(true)
    expect(isValidIsoDate(undefined)).toBe(false)
  })
})
