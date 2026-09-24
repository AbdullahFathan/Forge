import { describe, expect, it } from "vitest"

import { addDays, defaultCapacityRange } from "@/lib/date-range"

describe("defaultCapacityRange", () => {
  it("spans 28 inclusive days (4 weeks)", () => {
    const range = defaultCapacityRange()
    expect(addDays(range.from, 27)).toBe(range.to)
  })
})
