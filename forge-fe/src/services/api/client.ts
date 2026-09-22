import axios, {
  type AxiosError,
  type AxiosInstance,
  type InternalAxiosRequestConfig,
} from "axios"

import { env } from "@/lib/env"
import { useSession } from "@/lib/session"
import { getAccessToken } from "@/lib/storage"
import { RETURN_TO_KEY } from "@/lib/constants"
import { ApiError, type ApiErrorBody, type Envelope } from "@/types/api"

export type RetryConfig = InternalAxiosRequestConfig & { _retry?: boolean }

export type TokenStore = {
  get: () => string | null
  set: (token: string | null) => void
}

const fallbackError: ApiErrorBody = {
  code: "INTERNAL_ERROR",
  message: "Request failed",
}

export function unwrapEnvelope<T>(body: Envelope<T>, status = 400): T {
  if (!body?.success || body.data === undefined) {
    throw new ApiError(status, body?.error ?? fallbackError)
  }
  return body.data
}

export function toApiError(error: AxiosError<Envelope<unknown>>) {
  const status = error.response?.status ?? 0
  const body = error.response?.data
  if (body?.error) return new ApiError(status, body.error)
  if (error instanceof ApiError) return error
  return new ApiError(status || 500, {
    code: "INTERNAL_ERROR",
    message: error.message || "Request failed",
  })
}

function isAuthPath(url: string | undefined) {
  if (!url) return false
  return (
    url.includes("/auth/login") ||
    url.includes("/auth/refresh") ||
    url.includes("/auth/logout")
  )
}

export function createApiClient(options?: {
  baseURL?: string
  tokens?: TokenStore
  onSessionCleared?: () => void
  refresh?: () => Promise<string>
}) {
  const tokens = options?.tokens ?? {
    get: getAccessToken,
    set: (token) => {
      if (token) useSession.getState().setFromToken(token)
      else useSession.getState().clear()
    },
  }
  const baseURL = options?.baseURL ?? env.apiUrl

  const client = axios.create({
    baseURL,
    withCredentials: true,
  })

  client.interceptors.request.use((config) => {
    const token = tokens.get()
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  })

  let refreshPromise: Promise<string> | null = null

  async function refreshAccessToken() {
    if (!refreshPromise) {
      refreshPromise = (
        options?.refresh
          ? options.refresh()
          : axios
              .post<Envelope<{ accessToken: string }>>(
                `${baseURL}/auth/refresh`,
                {},
                { withCredentials: true },
              )
              .then((response) => unwrapEnvelope(response.data, response.status).accessToken)
      )
        .then((token) => {
          tokens.set(token)
          return token
        })
        .finally(() => {
          refreshPromise = null
        })
    }
    return refreshPromise
  }

  client.interceptors.response.use(
    (response) => {
      const body = response.data as Envelope<unknown> | undefined
      if (body && typeof body === "object" && "success" in body) {
        response.data = unwrapEnvelope(body, response.status)
      }
      return response
    },
    async (error: AxiosError<Envelope<unknown>>) => {
      const original = error.config as RetryConfig | undefined
      const status = error.response?.status
      if (
        status === 401 &&
        original &&
        !original._retry &&
        !isAuthPath(original.url)
      ) {
        original._retry = true
        try {
          const token = await refreshAccessToken()
          original.headers.set("Authorization", `Bearer ${token}`)
          return client(original)
        } catch (refreshError) {
          tokens.set(null)
          options?.onSessionCleared?.()
          return Promise.reject(
            refreshError instanceof ApiError
              ? refreshError
              : toApiError(error),
          )
        }
      }
      return Promise.reject(toApiError(error))
    },
  )

  return client
}

let client: AxiosInstance | null = null

export function getApiClient() {
  if (!client) {
    client = createApiClient({
      onSessionCleared: () => {
        if (typeof window === "undefined") return
        if (window.location.pathname === "/login") return
        const next = `${window.location.pathname}${window.location.search}`
        sessionStorage.setItem(RETURN_TO_KEY, next)
        window.location.assign("/login")
      },
    })
  }
  return client
}
