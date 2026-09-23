import { describe, expect, it } from "vitest"

import { holidaySchema } from "@/features/holidays/schema"

describe("holidaySchema", () => {
  const valid = { date: "2026-08-17", name: "Hari Kemerdekaan RI" }

  it("accepts a valid holiday", () => {
    expect(holidaySchema.safeParse(valid).success).toBe(true)
  })

  it("blocks a bad date format", () => {
    const result = holidaySchema.safeParse({ ...valid, date: "17-08-2026" })
    expect(result.success).toBe(false)
  })

  it("blocks an empty name", () => {
    const result = holidaySchema.safeParse({ ...valid, name: "   " })
    expect(result.success).toBe(false)
  })
})
