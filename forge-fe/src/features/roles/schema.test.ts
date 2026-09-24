import { describe, expect, it } from "vitest"

import { roleFormSchema } from "@/features/roles/schema"

describe("roleFormSchema", () => {
  const valid = {
    name: "Auditor",
    permissionCodes: ["audit.read"],
  }

  it("accepts a valid custom role", () => {
    expect(roleFormSchema.safeParse(valid).success).toBe(true)
  })

  it("blocks a short name", () => {
    const result = roleFormSchema.safeParse({ ...valid, name: "A" })
    expect(result.success).toBe(false)
  })

  it("blocks a blank name", () => {
    const result = roleFormSchema.safeParse({ ...valid, name: "  " })
    expect(result.success).toBe(false)
  })

  it("blocks an empty permission set", () => {
    const result = roleFormSchema.safeParse({ ...valid, permissionCodes: [] })
    expect(result.success).toBe(false)
  })
})
