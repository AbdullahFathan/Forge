import { useQuery } from "@tanstack/react-query"
import { useState } from "react"
import { Link } from "react-router"

import { ChartCard } from "@/components/common/chart-card"
import { EntityTable } from "@/components/common/entity-table"
import { PageHeader } from "@/components/common/page-header"
import { PriorityBadge } from "@/components/common/priority-badge"
import { StatusBadge } from "@/components/common/status-badge"
import { Label } from "@/components/ui/label"
import { Input } from "@/components/ui/input"
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { getWorkload } from "@/features/resources/api/capacity"
import { listUsers } from "@/features/users/api/users"
import { can } from "@/lib/auth"
import { PERMISSIONS } from "@/lib/constants"
import { useSession } from "@/lib/session"
import { queryKeys } from "@/services/query/query-keys"
import { ApiError } from "@/types/api"

const SELF = "self"

export function WorkloadPage() {
  const sessionUser = useSession((state) => state.user)
  const permissions = useSession((state) => state.permissions)
  const canPickUser =
    can(permissions, PERMISSIONS.userManage) || can(permissions, PERMISSIONS.capacityView)
  const canListUsers = can(permissions, PERMISSIONS.userManage)

  const [userId, setUserId] = useState(SELF)

  const users = useQuery({
    queryKey: queryKeys.users({ page: 1, pageSize: 100, picker: true }),
    queryFn: () => listUsers({ page: 1, pageSize: 100 }),
    enabled: canListUsers,
  })

  const targetId = userId === SELF ? undefined : userId

  const workload = useQuery({
    queryKey: queryKeys.workload(targetId),
    queryFn: () => getWorkload(targetId),
  })

  const data = workload.data
  const error =
    workload.error instanceof ApiError ? workload.error.message : workload.error ? "Could not load workload" : null

  return (
    <div className="flex flex-col gap-4">
      <PageHeader
        title="My workload"
        description="Projects, assigned tasks, and the next four weeks of planned load."
      />
      {canPickUser ? (
        canListUsers ? (
          <div className="flex max-w-sm flex-col gap-1">
            <Label>Person</Label>
            <Select value={userId} onValueChange={(value) => setUserId(value ?? SELF)}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  <SelectItem value={SELF}>Me{sessionUser?.name ? ` (${sessionUser.name})` : ""}</SelectItem>
                  {(users.data?.items ?? []).map((user) => (
                    <SelectItem key={user.id} value={user.id}>
                      {user.name}
                    </SelectItem>
                  ))}
                </SelectGroup>
              </SelectContent>
            </Select>
          </div>
        ) : (
          <div className="flex max-w-sm flex-col gap-1">
            <Label htmlFor="workload-user">User ID (optional)</Label>
            <Input
              id="workload-user"
              placeholder="Leave blank for yourself"
              value={userId === SELF ? "" : userId}
              onChange={(event) => setUserId(event.target.value.trim() || SELF)}
            />
          </div>
        )
      ) : null}
      <EntityTable
        columns={[
          {
            key: "project",
            header: "Project",
            cell: (row) => (
              <Link className="underline-offset-4 hover:underline" to={`/projects/${row.projectId}`}>
                {row.projectName ?? row.projectId}
              </Link>
            ),
          },
          {
            key: "percent",
            header: "%",
            cell: (row) => <span className="tabular-nums">{row.allocationPercent}%</span>,
          },
          { key: "role", header: "Role", cell: (row) => row.role },
          { key: "dates", header: "Period", cell: (row) => `${row.startDate} → ${row.endDate}` },
        ]}
        rows={data?.projects ?? []}
        rowKey={(row) => `${row.projectId}-${row.startDate}`}
        isLoading={workload.isLoading}
        error={error}
        emptyTitle="No project allocations"
        emptyDescription="You are not allocated to a project in this window."
        page={1}
        pageSize={Math.max(data?.projects.length ?? 1, 1)}
        totalItems={data?.projects.length ?? 0}
        onPageChange={() => undefined}
      />
      <EntityTable
        columns={[
          {
            key: "name",
            header: "Task",
            cell: (row) => (
              <Link className="underline-offset-4 hover:underline" to={`/projects/${row.projectId}`}>
                {row.name}
              </Link>
            ),
          },
          {
            key: "status",
            header: "Status",
            cell: (row) => <StatusBadge kind="task" value={row.status} />,
          },
          {
            key: "priority",
            header: "Priority",
            cell: (row) => <PriorityBadge value={row.priority} />,
          },
          { key: "due", header: "Due", cell: (row) => row.dueDate ?? "—" },
        ]}
        rows={data?.tasks ?? []}
        rowKey={(row) => row.id}
        isLoading={workload.isLoading}
        error={error}
        emptyTitle="No assigned tasks"
        page={1}
        pageSize={Math.max(data?.tasks.length ?? 1, 1)}
        totalItems={data?.tasks.length ?? 0}
        onPageChange={() => undefined}
      />
      <ChartCard
        title="Weekly load"
        description="Utilization for the next four weeks."
        data={(data?.weekSeries ?? []).map((bucket) => ({
          period: bucket.periodKey,
          value: Math.round(bucket.utilizationPercent),
        }))}
      />
    </div>
  )
}
