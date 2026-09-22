import { getApiClient } from "@/services/api/client"
import { endpoints } from "@/services/api/endpoints"
import type { Page } from "@/types/api"
import type { Department } from "@/types/common"

export async function listDepartments(page: number, pageSize: number) {
  const { data } = await getApiClient().get<Page<Department>>(endpoints.departments, {
    params: { page, pageSize },
  })
  return data
}

export async function createDepartment(name: string) {
  const { data } = await getApiClient().post<Department>(endpoints.departments, { name })
  return data
}

export async function patchDepartment(id: string, name: string) {
  const { data } = await getApiClient().patch<Department>(endpoints.department(id), { name })
  return data
}

export async function deleteDepartment(id: string) {
  const { data } = await getApiClient().delete<{ status: string }>(endpoints.department(id))
  return data
}
