import {
  AlertTriangleIcon,
  BellIcon,
  CalendarClockIcon,
  CheckCircle2Icon,
  CircleDotIcon,
  UserMinusIcon,
  UsersIcon,
} from "lucide-react"

import { formatRelativeTime } from "@/lib/relative-time"
import { NOTIFICATION_TYPE_LABELS, type NotificationPublic } from "@/types/notification"

const icons: Record<string, typeof BellIcon> = {
  TASK_ASSIGNED: UsersIcon,
  TASK_DUE_SOON: CalendarClockIcon,
  TASK_OVERDUE: AlertTriangleIcon,
  TASK_STATUS_CHANGED: CheckCircle2Icon,
  RESOURCE_OVERLOAD: AlertTriangleIcon,
  PROJECT_DUE_SOON: CalendarClockIcon,
  MEMBER_REMOVED: UserMinusIcon,
}

export function NotificationItem({
  item,
  onMarkRead,
}: {
  item: NotificationPublic
  onMarkRead?: (id: string) => void
}) {
  const Icon = icons[item.type] ?? CircleDotIcon
  const typeLabel = NOTIFICATION_TYPE_LABELS[item.type] ?? item.type

  return (
    <div className="flex items-start gap-3 rounded-lg border p-3">
      <span
        className={item.isRead ? "mt-1 size-2 shrink-0 rounded-full bg-transparent" : "mt-1 size-2 shrink-0 rounded-full bg-primary"}
        aria-hidden
      />
      <Icon className="mt-0.5 size-4 shrink-0 text-muted-foreground" />
      <div className="min-w-0 flex-1">
        <div className="flex flex-wrap items-baseline justify-between gap-2">
          <p className="text-sm font-medium">{item.title}</p>
          <time
            className="text-xs text-muted-foreground"
            dateTime={item.createdAt}
            title={new Date(item.createdAt).toUTCString()}
          >
            {formatRelativeTime(item.createdAt)}
          </time>
        </div>
        <p className="text-xs text-muted-foreground">{typeLabel}</p>
        {item.body ? <p className="mt-1 text-sm">{item.body}</p> : null}
        {!item.isRead && onMarkRead ? (
          <button
            type="button"
            className="mt-2 text-xs font-medium text-primary underline-offset-4 hover:underline"
            onClick={() => onMarkRead(item.id)}
          >
            Mark as read
          </button>
        ) : null}
      </div>
    </div>
  )
}
