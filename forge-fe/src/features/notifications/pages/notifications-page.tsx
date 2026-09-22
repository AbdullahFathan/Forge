import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { useState } from "react"
import { toast } from "sonner"

import { EmptyState } from "@/components/common/empty-state"
import { FilterBar } from "@/components/common/filter-bar"
import { NotificationItem } from "@/components/common/notification-item"
import { PageHeader } from "@/components/common/page-header"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { Button } from "@/components/ui/button"
import { Label } from "@/components/ui/label"
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Spinner } from "@/components/ui/spinner"
import { listNotifications, markAllNotificationsRead } from "@/features/notifications/api/notifications"
import { useMarkNotificationRead } from "@/features/notifications/hooks/use-mark-notification-read"
import { compactParams } from "@/lib/search-params"
import { DEFAULT_PAGE_SIZE, NOTIFICATION_TYPES } from "@/lib/constants"
import { queryKeys } from "@/services/query/query-keys"
import { ApiError } from "@/types/api"
import { NOTIFICATION_TYPE_LABELS } from "@/types/notification"

const ALL = "all"

export function NotificationsPage() {
  const queryClient = useQueryClient()
  const [page, setPage] = useState(1)
  const [type, setType] = useState(ALL)
  const [readFilter, setReadFilter] = useState(ALL)

  const params = compactParams({
    page,
    pageSize: DEFAULT_PAGE_SIZE,
    type: type === ALL ? undefined : type,
    unread: readFilter === "unread" ? true : readFilter === "read" ? false : undefined,
  })

  const list = useQuery({
    queryKey: queryKeys.notifications(params),
    queryFn: () =>
      listNotifications({
        page,
        pageSize: DEFAULT_PAGE_SIZE,
        ...(type !== ALL ? { type } : {}),
        ...(readFilter === "unread" ? { unread: true } : {}),
        ...(readFilter === "read" ? { unread: false } : {}),
      }),
  })

  const invalidate = () => {
    void queryClient.invalidateQueries({ queryKey: ["notifications"] })
  }

  const markOne = useMarkNotificationRead()

  const markAll = useMutation({
    mutationFn: markAllNotificationsRead,
    onSuccess: () => {
      invalidate()
      toast.success("All notifications marked as read")
    },
    onError: (error) => {
      toast.error(error instanceof ApiError ? error.message : "Could not mark all as read")
    },
  })

  const error =
    list.error instanceof ApiError ? list.error.message : list.error ? "Could not load notifications" : null
  const items = list.data?.items ?? []
  const total = list.data?.totalItems ?? 0
  const pageCount = Math.max(1, Math.ceil(total / DEFAULT_PAGE_SIZE))

  return (
    <div className="flex flex-col gap-4">
      <PageHeader
        title="Notifications"
        description="Your in-app alerts. Email delivery is toggled on your profile."
        actions={
          <Button
            type="button"
            variant="outline"
            disabled={markAll.isPending}
            onClick={() => markAll.mutate()}
          >
            {markAll.isPending ? <Spinner data-icon="inline-start" /> : null}
            Mark all read
          </Button>
        }
      />
      <FilterBar>
        <div className="flex flex-col gap-1">
          <Label>Type</Label>
          <Select
            value={type}
            onValueChange={(value) => {
              setType(value ?? ALL)
              setPage(1)
            }}
          >
            <SelectTrigger className="w-48">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectGroup>
                <SelectItem value={ALL}>All types</SelectItem>
                {NOTIFICATION_TYPES.map((code) => (
                  <SelectItem key={code} value={code}>
                    {NOTIFICATION_TYPE_LABELS[code]}
                  </SelectItem>
                ))}
              </SelectGroup>
            </SelectContent>
          </Select>
        </div>
        <div className="flex flex-col gap-1">
          <Label>Status</Label>
          <Select
            value={readFilter}
            onValueChange={(value) => {
              setReadFilter(value ?? ALL)
              setPage(1)
            }}
          >
            <SelectTrigger className="w-40">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectGroup>
                <SelectItem value={ALL}>All</SelectItem>
                <SelectItem value="unread">Unread</SelectItem>
                <SelectItem value="read">Read</SelectItem>
              </SelectGroup>
            </SelectContent>
          </Select>
        </div>
      </FilterBar>
      {error ? (
        <Alert variant="destructive">
          <AlertTitle>Could not load notifications</AlertTitle>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      ) : null}
      {!list.isLoading && items.length === 0 ? (
        <EmptyState title="No notifications" description="New events will show up here." />
      ) : (
        <div className="flex flex-col gap-2">
          {items.map((item) => (
            <NotificationItem
              key={item.id}
              item={item}
              onMarkRead={(id) => markOne.mutate(id)}
            />
          ))}
        </div>
      )}
      <div className="flex items-center justify-between gap-2">
        <p className="text-xs text-muted-foreground tabular-nums">{total} total</p>
        <div className="flex items-center gap-2">
          <Button
            type="button"
            variant="outline"
            size="sm"
            disabled={page <= 1 || list.isLoading}
            onClick={() => setPage(page - 1)}
          >
            Previous
          </Button>
          <span className="text-xs tabular-nums text-muted-foreground">
            {page} / {pageCount}
          </span>
          <Button
            type="button"
            variant="outline"
            size="sm"
            disabled={page >= pageCount || list.isLoading}
            onClick={() => setPage(page + 1)}
          >
            Next
          </Button>
        </div>
      </div>
    </div>
  )
}
