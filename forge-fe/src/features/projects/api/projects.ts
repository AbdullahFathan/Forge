import { getApiClient } from "@/services/api/client"
import { endpoints } from "@/services/api/endpoints"
import { compactParams } from "@/lib/search-params"
import type { Page } from "@/types/api"
import type { ProjectActivity, ProjectPublic, ProjectSummary } from "@/types/project"

export type ProjectListParams = {
  page: number
  pageSize: number
  status?: string
  departmentId?: string
  q?: string
}

export type ProjectWrite = {
  name: string
  description?: string
  status?: string
  priority?: string
  startDate: string
  targetEndDate: string
  ownerId: string
  departmentId?: string | null
  clearDepartment?: boolean
  tags: string[]
}

export async function listProjects(params: ProjectListParams) {
  const { data } = await getApiClient().get<Page<ProjectPublic>>(endpoints.projects, {
    params: compactParams(params),
  })
  return data
}

export async function getProject(id: string) {
  const { data } = await getApiClient().get<ProjectSummary>(endpoints.project(id))
  return data
}

export async function createProject(input: ProjectWrite) {
  const { data } = await getApiClient().post<ProjectPublic>(endpoints.projects, input)
  return data
}

export async function patchProject(id: string, input: Partial<ProjectWrite>) {
  const { data } = await getApiClient().patch<ProjectPublic>(endpoints.project(id), input)
  return data
}

export async function archiveProject(id: string) {
  const { data } = await getApiClient().delete<{ status: string }>(endpoints.project(id))
  return data
}

export async function listProjectActivity(id: string, page: number, pageSize: number) {
  const { data } = await getApiClient().get<Page<ProjectActivity>>(endpoints.projectActivity(id), {
    params: { page, pageSize },
  })
  return data
}
