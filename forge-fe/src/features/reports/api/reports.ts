import { getApiClient } from "@/services/api/client"
import { endpoints } from "@/services/api/endpoints"
import { compactParams } from "@/lib/search-params"
import type {
  ProjectStatusRow,
  ReportFilters,
  TaskCompletionRow,
  UtilizationRow,
} from "@/types/report"

export async function getProjectStatusReport(filters: ReportFilters) {
  const { data } = await getApiClient().get<ProjectStatusRow[]>(endpoints.reportsProjectStatus, {
    params: compactParams(filters),
  })
  return data
}

export async function getUtilizationReport(filters: ReportFilters) {
  const { data } = await getApiClient().get<UtilizationRow[]>(endpoints.reportsUtilization, {
    params: compactParams(filters),
  })
  return data
}

export async function getTaskCompletionReport(filters: ReportFilters) {
  const { data } = await getApiClient().get<TaskCompletionRow[]>(endpoints.reportsTaskCompletion, {
    params: compactParams(filters),
  })
  return data
}
