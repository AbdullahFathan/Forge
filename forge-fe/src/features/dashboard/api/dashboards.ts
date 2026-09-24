import { getApiClient } from "@/services/api/client"
import { endpoints } from "@/services/api/endpoints"
import type { DashboardKind } from "@/lib/auth"
import type {
  ExecutiveDashboard,
  MemberDashboard,
  ProjectManagerDashboard,
} from "@/types/dashboard"

export function usesDashboardApi(kind: DashboardKind) {
  return kind === "executive" || kind === "project-manager" || kind === "member"
}

export function dashboardPath(kind: DashboardKind) {
  if (kind === "executive") return endpoints.dashboardExecutive
  if (kind === "project-manager") return endpoints.dashboardProjectManager
  if (kind === "member") return endpoints.dashboardMember
  return null
}

export function dashboardWidgetLabels(kind: DashboardKind) {
  if (kind === "executive") return ["Active projects", "Late", "Org utilization"]
  if (kind === "project-manager") return ["Blocked", "Overdue", "Unassigned"]
  if (kind === "resource-manager") return ["Overload (14 days)", "Available"]
  return ["Assigned tasks", "This week’s load"]
}

export async function getDashboard(kind: DashboardKind) {
  const path = dashboardPath(kind)
  if (!path || kind === "resource-manager") {
    throw new Error("Resource manager dashboard does not use a dashboard API")
  }
  if (kind === "executive") {
    const { data } = await getApiClient().get<ExecutiveDashboard>(path)
    return data
  }
  if (kind === "project-manager") {
    const { data } = await getApiClient().get<ProjectManagerDashboard>(path)
    return data
  }
  const { data } = await getApiClient().get<MemberDashboard>(path)
  return data
}
