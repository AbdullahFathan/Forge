export type NotificationType =
  | "TASK_ASSIGNED"
  | "TASK_DUE_SOON"
  | "TASK_OVERDUE"
  | "TASK_STATUS_CHANGED"
  | "RESOURCE_OVERLOAD"
  | "PROJECT_DUE_SOON"
  | "MEMBER_REMOVED"

export type NotificationPublic = {
  id: string
  type: NotificationType | string
  title: string
  body: string
  isRead: boolean
  metadata: Record<string, unknown>
  entityId: string
  createdAt: string
}

export type UnreadCount = {
  unreadCount: number
}

export const NOTIFICATION_TYPE_LABELS: Record<string, string> = {
  TASK_ASSIGNED: "Task assigned",
  TASK_DUE_SOON: "Due soon",
  TASK_OVERDUE: "Overdue",
  TASK_STATUS_CHANGED: "Status changed",
  RESOURCE_OVERLOAD: "Over-allocated",
  PROJECT_DUE_SOON: "Project due soon",
  MEMBER_REMOVED: "Removed from project",
}
