import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { renderHook, waitFor } from "@testing-library/react"
import { createElement, type ReactNode } from "react"
import { beforeEach, describe, expect, it, vi } from "vitest"

import { markNotificationRead } from "@/features/notifications/api/notifications"
import { queryKeys } from "@/services/query/query-keys"
import { ApiError, type Page } from "@/types/api"
import type { NotificationPublic, UnreadCount } from "@/types/notification"

import { useMarkNotificationRead } from "./use-mark-notification-read"

vi.mock("@/features/notifications/api/notifications", () => ({
  markNotificationRead: vi.fn(),
}))

vi.mock("sonner", () => ({
  toast: { error: vi.fn(), success: vi.fn() },
}))

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

function wrapper(client: QueryClient) {
  return function Wrapper({ children }: { children: ReactNode }) {
    return createElement(QueryClientProvider, { client }, children)
  }
}

describe("useMarkNotificationRead", () => {
  const markRead = markNotificationRead as unknown as ReturnType<typeof vi.fn>

  beforeEach(() => {
    markRead.mockReset()
  })

  it("rolls back isRead and unread count when the API errors", async () => {
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

    markRead.mockRejectedValue(
      new ApiError(500, { code: "INTERNAL_ERROR", message: "mark failed" }),
    )

    const { result } = renderHook(() => useMarkNotificationRead(), { wrapper: wrapper(client) })

    await expect(result.current.mutateAsync("n1")).rejects.toBeInstanceOf(ApiError)

    await waitFor(() => {
      const list = client.getQueryData<Page<NotificationPublic>>(listKey)
      const unread = client.getQueryData<UnreadCount>(queryKeys.notificationsUnread)
      expect(list?.items[0]?.isRead).toBe(false)
      expect(unread?.unreadCount).toBe(1)
    })
  })
})
