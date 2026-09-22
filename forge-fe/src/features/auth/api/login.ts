import { getApiClient } from "@/services/api/client"
import { endpoints } from "@/services/api/endpoints"
import type { LoginResult } from "@/types/user"

export async function login(input: { email: string; password: string }) {
  const { data } = await getApiClient().post<LoginResult>(endpoints.login, input)
  return data
}
