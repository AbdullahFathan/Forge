import { getApiClient } from "@/services/api/client"
import { endpoints } from "@/services/api/endpoints"
import type { RoleSummary } from "@/types/common"

export async function listRoles() {
  const { data } = await getApiClient().get<RoleSummary[]>(endpoints.roles)
  return data
}
