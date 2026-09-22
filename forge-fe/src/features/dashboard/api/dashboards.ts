import { getApiClient } from "@/services/api/client"
import { endpoints } from "@/services/api/endpoints"
import type { DashboardKind } from "@/lib/auth"
import type {
  ExecutiveDashboard,
  MemberDashboard,
  ProjectManagerDashboard,
} from "@/types/dashboard"

export function dashboardPath(kind: DashboardKind) {
  if (kind === "executive") return endpoints.dashboardExecutive
  if (kind === "project-manager") return endpoints.dashboardProjectManager
  return endpoints.dashboardMember
}

export function dashboardWidgetLabels(kind: DashboardKind) {
  if (kind === "executive") return ["Active projects", "Late", "Org utilization"]
  if (kind === "project-manager") return ["Blocked", "Overdue", "Unassigned"]
  return ["Assigned tasks", "This week’s load"]
}

export async function getDashboard(kind: DashboardKind) {
  const path = dashboardPath(kind)
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
