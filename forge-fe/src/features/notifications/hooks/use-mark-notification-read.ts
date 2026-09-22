import { useMutation, useQueryClient } from "@tanstack/react-query"
import { toast } from "sonner"

import { markNotificationRead } from "@/features/notifications/api/notifications"
import { queryKeys } from "@/services/query/query-keys"
import { ApiError, type Page } from "@/types/api"
import type { NotificationPublic, UnreadCount } from "@/types/notification"

type Snapshot = [readonly unknown[], unknown][]

function isNotificationListKey(queryKey: readonly unknown[]) {
  return queryKey[0] === "notifications" && queryKey[1] !== "unread-count"
}

export function useMarkNotificationRead() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: markNotificationRead,
    onMutate: async (id) => {
      await queryClient.cancelQueries({ queryKey: ["notifications"] })
      const previous: Snapshot = queryClient.getQueriesData({ queryKey: ["notifications"] })

      queryClient.setQueriesData<Page<NotificationPublic>>(
        { predicate: (query) => isNotificationListKey(query.queryKey) },
        (old) => {
          if (!old?.items) return old
          return {
            ...old,
            items: old.items.map((item) => (item.id === id ? { ...item, isRead: true } : item)),
          }
        },
      )

      queryClient.setQueryData<UnreadCount>(queryKeys.notificationsUnread, (old) => {
        if (!old) return old
        return { unreadCount: Math.max(0, old.unreadCount - 1) }
      })

      return { previous }
    },
    onError: (error, _id, context) => {
      context?.previous.forEach(([key, data]) => {
        queryClient.setQueryData(key, data)
      })
      toast.error(error instanceof ApiError ? error.message : "Could not mark as read")
    },
    onSettled: () => {
      void queryClient.invalidateQueries({ queryKey: ["notifications"] })
    },
  })
}
