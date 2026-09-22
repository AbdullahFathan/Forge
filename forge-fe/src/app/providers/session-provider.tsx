import { useEffect } from "react"
import axios from "axios"

import { Loader } from "@/components/common/loader"
import { env } from "@/lib/env"
import { getAccessToken } from "@/lib/storage"
import { useSession } from "@/lib/session"
import { endpoints } from "@/services/api/endpoints"
import { unwrapEnvelope } from "@/services/api/client"
import type { Envelope } from "@/types/api"
import type { LoginResult } from "@/types/user"

export function SessionProvider({ children }: { children: React.ReactNode }) {
  const bootstrapped = useSession((state) => state.bootstrapped)
  const setFromToken = useSession((state) => state.setFromToken)
  const setBootstrapped = useSession((state) => state.setBootstrapped)

  useEffect(() => {
    let cancelled = false

    async function boot() {
      if (getAccessToken()) {
        if (!cancelled) setBootstrapped(true)
        return
      }
      try {
        const response = await axios.post<Envelope<LoginResult>>(
          `${env.apiUrl}${endpoints.refresh}`,
          {},
          { withCredentials: true },
        )
        const data = unwrapEnvelope(response.data, response.status)
        if (!cancelled) setFromToken(data.accessToken, data.user)
      } catch {
        if (!cancelled) useSession.getState().clear()
      } finally {
        if (!cancelled) setBootstrapped(true)
      }
    }

    void boot()
    return () => {
      cancelled = true
    }
  }, [setBootstrapped, setFromToken])

  if (!bootstrapped) {
    return (
      <div className="flex min-h-svh items-center justify-center">
        <Loader label="Restoring session" />
      </div>
    )
  }

  return children
}
