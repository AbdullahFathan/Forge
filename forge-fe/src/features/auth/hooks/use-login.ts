import { useMutation } from "@tanstack/react-query"
import { useNavigate } from "react-router"

import { consumeReturnTo } from "@/lib/return-to"
import { useSession } from "@/lib/session"
import { login } from "@/features/auth/api/login"

export function useLogin() {
  const navigate = useNavigate()
  const setFromToken = useSession((state) => state.setFromToken)

  return useMutation({
    mutationFn: login,
    onSuccess: (result) => {
      setFromToken(result.accessToken, result.user)
      navigate(consumeReturnTo(), { replace: true })
    },
  })
}
