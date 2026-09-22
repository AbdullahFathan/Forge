import { describe, expect, it } from "vitest"

import { OverAllocationBanner } from "@/features/resources/components/over-allocation-banner"

describe("OverAllocationBanner", () => {
  it("renders when the save payload is over-allocated", () => {
    const node = OverAllocationBanner({ overAllocated: true, warnings: ["OVER_ALLOCATED"] })
    expect(node).not.toBeNull()
  })

  it("does not treat a clean save as an error banner", () => {
    const node = OverAllocationBanner({ overAllocated: false, warnings: [] })
    expect(node).toBeNull()
  })
})
