import { getApiClient } from "@/services/api/client"
import { endpoints } from "@/services/api/endpoints"
import type { PublicUser } from "@/types/user"

export type ProfilePatch = {
  name?: string
  password?: string
  currentPassword?: string
  emailNotificationsEnabled?: boolean
}

export async function getMe() {
  const { data } = await getApiClient().get<PublicUser>(endpoints.me)
  return data
}

export async function patchMe(input: ProfilePatch) {
  const { data } = await getApiClient().patch<PublicUser>(endpoints.me, input)
  return data
}
