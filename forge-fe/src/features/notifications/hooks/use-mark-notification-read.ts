import { useMutation, useQueryClient } from "@tanstack/react-query"
import { toast } from "sonner"

import { markNotificationRead } from "@/features/notifications/api/notifications"
import { ApiError } from "@/types/api"

import {
  markNotificationReadInCache,
  restoreNotificationQueries,
  snapshotNotificationQueries,
} from "./mark-read-cache"

export function useMarkNotificationRead() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: markNotificationRead,
    onMutate: async (id) => {
      await queryClient.cancelQueries({ queryKey: ["notifications"] })
      const previous = snapshotNotificationQueries(queryClient)
      markNotificationReadInCache(queryClient, id)
      return { previous }
    },
    onError: (error, _id, context) => {
      restoreNotificationQueries(queryClient, context?.previous)
      toast.error(error instanceof ApiError ? error.message : "Could not mark as read")
    },
    onSettled: () => {
      void queryClient.invalidateQueries({ queryKey: ["notifications"] })
    },
  })
}
