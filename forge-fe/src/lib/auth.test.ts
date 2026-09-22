import { describe, expect, it } from "vitest"

import { can, visibleNav } from "@/lib/auth"
import { PERMISSIONS } from "@/lib/constants"

describe("can", () => {
  it("allows a held permission", () => {
    expect(can([PERMISSIONS.userManage], PERMISSIONS.userManage)).toBe(true)
  })

  it("denies a missing permission", () => {
    expect(can([PERMISSIONS.taskManage], PERMISSIONS.userManage)).toBe(false)
  })
})

describe("visibleNav", () => {
  it("hides admin items from a member", () => {
    const labels = visibleNav([PERMISSIONS.taskManage]).map((item) => item.label)
    expect(labels).toContain("Dashboard")
    expect(labels).toContain("Projects")
    expect(labels).not.toContain("Users")
    expect(labels).not.toContain("Audit log")
    expect(labels).not.toContain("Resources")
    expect(labels).not.toContain("Reports")
  })

  it("shows admin items when permissions match", () => {
    const labels = visibleNav([
      PERMISSIONS.userManage,
      PERMISSIONS.auditRead,
      PERMISSIONS.capacityView,
      PERMISSIONS.reportExport,
    ]).map((item) => item.label)
    expect(labels).toEqual(expect.arrayContaining(["Users", "Audit log", "Resources", "Reports"]))
  })

  it("shows Resources for a project manager with allocate only", () => {
    const labels = visibleNav([PERMISSIONS.resourceAllocate, PERMISSIONS.projectCreate]).map(
      (item) => item.label,
    )
    expect(labels).toContain("Resources")
    expect(labels).not.toContain("Reports")
    expect(labels).not.toContain("Audit log")
  })

  it("shows Resources for a resource manager", () => {
    const labels = visibleNav([PERMISSIONS.capacityView, PERMISSIONS.resourceAllocate]).map(
      (item) => item.label,
    )
    expect(labels).toContain("Resources")
  })
})
