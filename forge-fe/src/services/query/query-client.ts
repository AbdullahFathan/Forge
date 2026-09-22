import { QueryClient } from "@tanstack/react-query"

import { ApiError } from "@/types/api"

export function createQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: {
        retry: (failureCount, error) => {
          if (error instanceof ApiError && error.status >= 400 && error.status < 500) {
            return false
          }
          return failureCount < 1
        },
        refetchOnWindowFocus: false,
      },
    },
  })
}

export const queryClient = createQueryClient()
