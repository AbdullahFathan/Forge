import { getApiClient } from "@/services/api/client"
import { endpoints } from "@/services/api/endpoints"

export async function logout() {
  const { data } = await getApiClient().post<{ status: string }>(endpoints.logout)
  return data
}
