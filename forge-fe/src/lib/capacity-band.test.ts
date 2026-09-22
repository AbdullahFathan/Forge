import { describe, expect, it } from "vitest"

import {
  bandFromUtilization,
  isOverAllocated,
  sumUtilization,
} from "@/lib/capacity-band"

describe("capacity bands", () => {
  it("maps 50+50+10 overlapping percents to the over band", () => {
    const total = sumUtilization([50, 50, 10])
    expect(total).toBe(110)
    expect(bandFromUtilization(total)).toBe("RED")
  })

  it("treats OVER_ALLOCATED warnings as over-allocation", () => {
    expect(isOverAllocated(false, ["OVER_ALLOCATED"])).toBe(true)
    expect(isOverAllocated(true, [])).toBe(true)
    expect(isOverAllocated(false, [])).toBe(false)
  })

  it("keeps under 80% in the available band", () => {
    expect(bandFromUtilization(79)).toBe("GREEN")
    expect(bandFromUtilization(80)).toBe("YELLOW")
    expect(bandFromUtilization(100)).toBe("YELLOW")
  })
})
