import { create } from "zustand"

import { decodeAccessToken } from "@/lib/jwt"
import { setAccessToken } from "@/lib/storage"
import type { PublicUser } from "@/types/user"

type SessionState = {
  user: PublicUser | null
  permissions: string[]
  role: string | null
  bootstrapped: boolean
  setFromToken: (token: string, user?: PublicUser | null) => void
  setUser: (user: PublicUser) => void
  clear: () => void
  setBootstrapped: (value: boolean) => void
}

export const useSession = create<SessionState>((set) => ({
  user: null,
  permissions: [],
  role: null,
  bootstrapped: false,
  setFromToken: (token, user) => {
    setAccessToken(token)
    const claims = decodeAccessToken(token)
    set((state) => ({
      permissions: claims.perms ?? [],
      role: claims.role ?? state.role,
      user: user === undefined ? state.user : user,
    }))
  },
  setUser: (user) => set({ user, role: user.roleCode }),
  clear: () => {
    setAccessToken(null)
    set({ user: null, permissions: [], role: null })
  },
  setBootstrapped: (value) => set({ bootstrapped: value }),
}))
