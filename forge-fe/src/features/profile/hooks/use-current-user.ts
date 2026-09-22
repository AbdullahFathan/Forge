import { useQuery } from "@tanstack/react-query"

import { getMe } from "@/features/profile/api/profile"
import { getAccessToken } from "@/lib/storage"
import { useSession } from "@/lib/session"
import { queryKeys } from "@/services/query/query-keys"

export function useCurrentUser() {
  const setUser = useSession((state) => state.setUser)
  return useQuery({
    queryKey: queryKeys.me,
    enabled: Boolean(getAccessToken()),
    queryFn: async () => {
      const user = await getMe()
      setUser(user)
      return user
    },
  })
}
