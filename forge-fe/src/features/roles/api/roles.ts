import { getApiClient } from "@/services/api/client"
import { endpoints } from "@/services/api/endpoints"
import type { Permission, RoleSummary } from "@/types/common"

export async function listRoles() {
  const { data } = await getApiClient().get<RoleSummary[]>(endpoints.roles)
  return data
}

export async function listPermissions() {
  const { data } = await getApiClient().get<Permission[]>(endpoints.permissions)
  return data
}

export async function createRole(input: { name: string; permissionCodes: string[] }) {
  const { data } = await getApiClient().post<RoleSummary>(endpoints.roles, input)
  return data
}

export async function patchRole(
  id: string,
  input: { name: string; permissionCodes: string[] },
) {
  const { data } = await getApiClient().patch<RoleSummary>(endpoints.role(id), input)
  return data
}

export async function deleteRole(id: string) {
  const { data } = await getApiClient().delete<{ status: string }>(endpoints.role(id))
  return data
}
