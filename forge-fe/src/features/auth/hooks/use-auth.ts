import { useMutation, useQueryClient } from "@tanstack/react-query"
import { useNavigate } from "react-router"

import { logout } from "@/features/auth/api/logout"
import { useSession } from "@/lib/session"

export function useAuth() {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const user = useSession((state) => state.user)
  const permissions = useSession((state) => state.permissions)
  const role = useSession((state) => state.role)
  const clear = useSession((state) => state.clear)

  const logoutMutation = useMutation({
    mutationFn: logout,
    onSettled: () => {
      clear()
      queryClient.clear()
      navigate("/login", { replace: true })
    },
  })

  return {
    user,
    permissions,
    role,
    logout: () => logoutMutation.mutate(),
    isLoggingOut: logoutMutation.isPending,
  }
}
