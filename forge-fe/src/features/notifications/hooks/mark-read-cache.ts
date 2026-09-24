import type { QueryClient } from "@tanstack/react-query"

import { queryKeys } from "@/services/query/query-keys"
import type { Page } from "@/types/api"
import type { NotificationPublic, UnreadCount } from "@/types/notification"

export type NotificationQuerySnapshot = [readonly unknown[], unknown][]

export function isNotificationListKey(queryKey: readonly unknown[]) {
  return queryKey[0] === "notifications" && queryKey[1] !== "unread-count"
}

export function snapshotNotificationQueries(queryClient: QueryClient): NotificationQuerySnapshot {
  return queryClient.getQueriesData({ queryKey: ["notifications"] })
}

export function markNotificationReadInCache(queryClient: QueryClient, id: string) {
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
}

export function restoreNotificationQueries(
  queryClient: QueryClient,
  previous: NotificationQuerySnapshot | undefined,
) {
  previous?.forEach(([key, data]) => {
    queryClient.setQueryData(key, data)
  })
}
