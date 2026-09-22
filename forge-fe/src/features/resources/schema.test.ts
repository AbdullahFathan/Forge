import { describe, expect, it } from "vitest"

import { allocationFormSchema } from "@/features/resources/schema"

describe("allocationFormSchema", () => {
  const valid = {
    userId: "11111111-1111-4111-8111-111111111111",
    projectId: "22222222-2222-4222-8222-222222222222",
    allocationPercent: 50,
    startDate: "2026-01-01",
    endDate: "2026-01-31",
    role: "MEMBER" as const,
  }

  it("accepts a valid allocation", () => {
    expect(allocationFormSchema.safeParse(valid).success).toBe(true)
  })

  it("blocks end before start", () => {
    const result = allocationFormSchema.safeParse({ ...valid, endDate: "2025-12-01" })
    expect(result.success).toBe(false)
  })

  it("blocks percent over 100", () => {
    const result = allocationFormSchema.safeParse({ ...valid, allocationPercent: 110 })
    expect(result.success).toBe(false)
  })
})
