import { QueryClient } from "@tanstack/react-query"
import { describe, expect, it } from "vitest"

import { queryKeys } from "@/services/query/query-keys"
import type { Page } from "@/types/api"
import type { NotificationPublic, UnreadCount } from "@/types/notification"

import {
  markNotificationReadInCache,
  restoreNotificationQueries,
  snapshotNotificationQueries,
} from "./mark-read-cache"

const unreadItem: NotificationPublic = {
  id: "n1",
  type: "TASK_ASSIGNED",
  title: "Assigned",
  body: "A task was assigned to you",
  isRead: false,
  metadata: {},
  entityId: "t1",
  createdAt: "2026-09-22T00:00:00Z",
}

const listKey = queryKeys.notifications({ page: 1, pageSize: 20 })

function seedClient() {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  })
  client.setQueryData<Page<NotificationPublic>>(listKey, {
    items: [unreadItem],
    page: 1,
    pageSize: 20,
    totalItems: 1,
  })
  client.setQueryData<UnreadCount>(queryKeys.notificationsUnread, { unreadCount: 1 })
  return client
}

describe("useMarkNotificationRead", () => {
  it("rolls back isRead and unread count when the API errors", () => {
    const client = seedClient()
    const previous = snapshotNotificationQueries(client)

    markNotificationReadInCache(client, "n1")

    expect(client.getQueryData<Page<NotificationPublic>>(listKey)?.items[0]?.isRead).toBe(true)
    expect(client.getQueryData<UnreadCount>(queryKeys.notificationsUnread)?.unreadCount).toBe(0)

    restoreNotificationQueries(client, previous)

    expect(client.getQueryData<Page<NotificationPublic>>(listKey)?.items[0]?.isRead).toBe(false)
    expect(client.getQueryData<UnreadCount>(queryKeys.notificationsUnread)?.unreadCount).toBe(1)
  })
})
