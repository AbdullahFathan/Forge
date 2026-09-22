import { getApiClient } from "@/services/api/client"
import { endpoints } from "@/services/api/endpoints"
import type { Page } from "@/types/api"
import type { ProjectMember } from "@/types/project"

export async function listProjectMembers(projectId: string, page: number, pageSize: number) {
  const { data } = await getApiClient().get<Page<ProjectMember>>(endpoints.projectMembers(projectId), {
    params: { page, pageSize },
  })
  return data
}

export async function addProjectMember(projectId: string, userId: string, role: string) {
  const { data } = await getApiClient().post<ProjectMember>(endpoints.projectMembers(projectId), {
    userId,
    role,
  })
  return data
}

export async function removeProjectMember(projectId: string, userId: string) {
  const { data } = await getApiClient().delete<{ status: string }>(
    endpoints.projectMember(projectId, userId),
  )
  return data
}
