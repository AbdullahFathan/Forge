import { describe, expect, it } from "vitest"

import { loginSchema } from "@/features/auth/schema"

describe("loginSchema", () => {
  it("blocks an invalid email", () => {
    const result = loginSchema.safeParse({ email: "not-an-email", password: "secret" })
    expect(result.success).toBe(false)
  })

  it("accepts a valid login", () => {
    const result = loginSchema.safeParse({ email: "admin@workspace.local", password: "secret" })
    expect(result.success).toBe(true)
  })
})
