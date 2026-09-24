import { describe, expect, it } from "vitest"

import { can, dashboardKind, visibleNav } from "@/lib/auth"
import {
  PERMISSIONS,
  ROLE_ADMIN,
  ROLE_PROJECT_MANAGER,
  ROLE_RESOURCE_MANAGER,
  ROLE_SUPER_ADMIN,
} from "@/lib/constants"

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
    expect(labels).not.toContain("Holidays")
    expect(labels).toContain("Notifications")
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

  it("shows Holidays when department.manage is held", () => {
    const labels = visibleNav([PERMISSIONS.departmentManage]).map((item) => item.label)
    expect(labels).toContain("Holidays")
    expect(labels).toContain("Departments")
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

  it("shows Reports but not Audit for a resource manager with export", () => {
    const labels = visibleNav([
      PERMISSIONS.capacityView,
      PERMISSIONS.resourceAllocate,
      PERMISSIONS.reportExport,
    ]).map((item) => item.label)
    expect(labels).toContain("Reports")
    expect(labels).not.toContain("Audit log")
    expect(labels).not.toContain("Users")
  })
})

describe("dashboardKind", () => {
  it("maps admin to executive, RM to resource-manager, others to member", () => {
    expect(dashboardKind(ROLE_SUPER_ADMIN)).toBe("executive")
    expect(dashboardKind(ROLE_ADMIN)).toBe("executive")
    expect(dashboardKind(ROLE_PROJECT_MANAGER)).toBe("project-manager")
    expect(dashboardKind(ROLE_RESOURCE_MANAGER)).toBe("resource-manager")
    expect(dashboardKind(null)).toBe("member")
  })
})
