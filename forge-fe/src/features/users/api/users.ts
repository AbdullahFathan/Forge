import { getApiClient } from "@/services/api/client"
import { endpoints } from "@/services/api/endpoints"
import type { Page } from "@/types/api"
import type { PublicUser } from "@/types/user"

export type UserListParams = {
  page: number
  pageSize: number
  departmentId?: string
  roleId?: string
}

export type UserWrite = {
  name: string
  email?: string
  password?: string
  roleId: string
  departmentId?: string | null
  clearDepartment?: boolean
  capacityHoursPerDay: number
  isActive: boolean
  skills: string[]
}

export async function listUsers(params: UserListParams) {
  const { data } = await getApiClient().get<Page<PublicUser>>(endpoints.users, { params })
  return data
}

export async function createUser(input: UserWrite) {
  const { data } = await getApiClient().post<PublicUser>(endpoints.users, input)
  return data
}

export async function patchUser(id: string, input: Partial<UserWrite>) {
  const { data } = await getApiClient().patch<PublicUser>(endpoints.user(id), input)
  return data
}

export async function deleteUser(id: string) {
  const { data } = await getApiClient().delete<{ status: string }>(endpoints.user(id))
  return data
}
