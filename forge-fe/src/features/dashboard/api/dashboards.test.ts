import { describe, expect, it } from "vitest"

import { dashboardKind } from "@/lib/auth"
import {
  ROLE_ADMIN,
  ROLE_MEMBER,
  ROLE_PROJECT_MANAGER,
  ROLE_RESOURCE_MANAGER,
  ROLE_SUPER_ADMIN,
} from "@/lib/constants"
import { endpoints } from "@/services/api/endpoints"

import { dashboardPath, dashboardWidgetLabels } from "./dashboards"

describe("dashboardPath", () => {
  it("sends Super Admin and Admin to executive only", () => {
    expect(dashboardPath(dashboardKind(ROLE_SUPER_ADMIN))).toBe(endpoints.dashboardExecutive)
    expect(dashboardPath(dashboardKind(ROLE_ADMIN))).toBe(endpoints.dashboardExecutive)
  })

  it("sends Project Manager to the PM dashboard only", () => {
    expect(dashboardPath(dashboardKind(ROLE_PROJECT_MANAGER))).toBe(endpoints.dashboardProjectManager)
  })

  it("never sends Member or Resource Manager to executive", () => {
    for (const role of [ROLE_RESOURCE_MANAGER, ROLE_MEMBER, null]) {
      const path = dashboardPath(dashboardKind(role))
      expect(path).toBe(endpoints.dashboardMember)
      expect(path).not.toBe(endpoints.dashboardExecutive)
    }
  })
})

describe("dashboardWidgetLabels", () => {
  it("shows executive summary widgets", () => {
    expect(dashboardWidgetLabels("executive")).toEqual(
      expect.arrayContaining(["Active projects", "Late", "Org utilization"]),
    )
  })

  it("shows PM attention widgets", () => {
    expect(dashboardWidgetLabels("project-manager")).toEqual(
      expect.arrayContaining(["Blocked", "Overdue", "Unassigned"]),
    )
  })

  it("shows member workload widgets, not executive gauges", () => {
    const labels = dashboardWidgetLabels("member")
    expect(labels).toEqual(expect.arrayContaining(["Assigned tasks", "This week’s load"]))
    expect(labels).not.toContain("Org utilization")
    expect(labels).not.toContain("Late")
  })
})
