import { getApiClient } from "@/services/api/client"
import { endpoints } from "@/services/api/endpoints"
import { compactParams } from "@/lib/search-params"
import type { Page } from "@/types/api"
import type { NotificationPublic, UnreadCount } from "@/types/notification"

export type NotificationListParams = {
  page: number
  pageSize: number
  unread?: boolean
  type?: string
}

export async function listNotifications(params: NotificationListParams) {
  const { data } = await getApiClient().get<Page<NotificationPublic>>(endpoints.notifications, {
    params: compactParams(params),
  })
  return data
}

export async function getUnreadCount() {
  const { data } = await getApiClient().get<UnreadCount>(endpoints.notificationsUnreadCount)
  return data
}

export async function markNotificationRead(id: string) {
  const { data } = await getApiClient().patch<{ status: string }>(endpoints.notificationRead(id))
  return data
}

export async function markAllNotificationsRead() {
  const { data } = await getApiClient().post<{ status: string }>(endpoints.notificationsReadAll)
  return data
}
