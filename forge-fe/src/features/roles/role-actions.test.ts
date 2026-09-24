import { describe, expect, it } from "vitest"

import { roleActionsVisible } from "@/features/roles/role-actions"

describe("roleActionsVisible", () => {
  it("hides edit and delete for system roles", () => {
    expect(roleActionsVisible({ isSystem: true })).toBe(false)
  })

  it("shows edit and delete for custom roles", () => {
    expect(roleActionsVisible({ isSystem: false })).toBe(true)
  })
})
