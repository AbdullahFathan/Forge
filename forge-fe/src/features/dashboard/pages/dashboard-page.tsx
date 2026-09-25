import { Link } from "react-router"

import { ChartCard } from "@/components/common/chart-card"
import { EmptyState } from "@/components/common/empty-state"
import { EntityTable } from "@/components/common/entity-table"
import { JsonDiff } from "@/components/common/json-diff"
import { NotificationItem } from "@/components/common/notification-item"
import { PageHeader } from "@/components/common/page-header"
import { PriorityBadge } from "@/components/common/priority-badge"
import { StatusBadge } from "@/components/common/status-badge"
import { CapacityCell } from "@/components/common/capacity-cell"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { buttonVariants } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { getDashboard, usesDashboardApi } from "@/features/dashboard/api/dashboards"
import { ResourceManagerView } from "@/features/dashboard/components/resource-manager-view"
import { dashboardKind } from "@/lib/auth"
import { useSession } from "@/lib/session"
import { queryKeys } from "@/services/query/query-keys"
import { ApiError } from "@/types/api"
import type {
  ExecutiveDashboard,
  MemberDashboard,
  ProjectManagerDashboard,
} from "@/types/dashboard"
import { useQuery } from "@tanstack/react-query"

function StatCard({ label, value }: { label: string; value: string }) {
  return (
    <Card size="sm">
      <CardHeader>
        <CardTitle className="text-muted-foreground">{label}</CardTitle>
      </CardHeader>
      <CardContent>
        <p className="text-2xl font-semibold tabular-nums">{value}</p>
      </CardContent>
    </Card>
  )
}

function DashboardSkeleton() {
  return (
    <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
      {Array.from({ length: 4 }, (_, index) => (
        <Skeleton key={index} className="h-24 w-full" />
      ))}
    </div>
  )
}

function ExecutiveView({ data }: { data: ExecutiveDashboard }) {
  return (
    <>
      <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <StatCard label="Active projects" value={String(data.activeProjects)} />
        <StatCard label="Completed" value={String(data.completedProjects)} />
        <StatCard label="Late" value={String(data.lateProjects)} />
        <StatCard label="Org utilization" value={`${Math.round(data.orgUtilization)}%`} />
      </div>
      <ChartCard
        title="Projects completed"
        description="Monthly completed trend"
        unit=""
        data={(data.monthlyCompleted ?? []).map((point) => ({
          period: point.month,
          value: point.count,
        }))}
      />
      <EntityTable
        columns={[
          {
            key: "name",
            header: "Due soon",
            cell: (row) => (
              <Link className="underline-offset-4 hover:underline" to={`/projects/${row.id}`}>
                {row.name}
              </Link>
            ),
          },
          {
            key: "status",
            header: "Status",
            cell: (row) => <StatusBadge kind="project" value={row.status} />,
          },
          { key: "end", header: "Target", cell: (row) => row.targetEndDate },
        ]}
        rows={data.dueSoon ?? []}
        rowKey={(row) => row.id}
        emptyTitle="No projects due soon"
        page={1}
        pageSize={Math.max(data.dueSoon?.length ?? 1, 1)}
        totalItems={data.dueSoon?.length ?? 0}
        onPageChange={() => undefined}
      />
      <EntityTable
        columns={[
          { key: "name", header: "Top utilization", cell: (row) => row.name },
          {
            key: "util",
            header: "%",
            cell: (row) => <CapacityCell percent={row.utilizationPercent} band={row.band} />,
          },
        ]}
        rows={data.topUtilization ?? []}
        rowKey={(row) => row.userId}
        emptyTitle="No utilization data"
        page={1}
        pageSize={Math.max(data.topUtilization?.length ?? 1, 1)}
        totalItems={data.topUtilization?.length ?? 0}
        onPageChange={() => undefined}
      />
    </>
  )
}

function ProjectManagerView({ data }: { data: ProjectManagerDashboard }) {
  return (
    <>
      <EntityTable
        columns={[
          {
            key: "name",
            header: "Project",
            cell: (row) => (
              <Link className="underline-offset-4 hover:underline" to={`/projects/${row.id}`}>
                {row.name}
              </Link>
            ),
          },
          {
            key: "status",
            header: "Status",
            cell: (row) => <StatusBadge kind="project" value={row.status} />,
          },
          {
            key: "progress",
            header: "Progress",
            cell: (row) => <span className="tabular-nums">{Math.round(row.completionPercent)}%</span>,
          },
          { key: "end", header: "Target", cell: (row) => row.targetEndDate },
        ]}
        rows={data.projects ?? []}
        rowKey={(row) => row.id}
        emptyTitle="No owned projects"
        page={1}
        pageSize={Math.max(data.projects?.length ?? 1, 1)}
        totalItems={data.projects?.length ?? 0}
        onPageChange={() => undefined}
      />
      <div className="rounded-lg ring-1 ring-destructive/30">
        <EntityTable
          columns={[
            {
              key: "name",
              header: "Blocked",
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
            { key: "due", header: "Due", cell: (row) => row.dueDate ?? "—" },
          ]}
          rows={data.blocked ?? []}
          rowKey={(row) => row.id}
          emptyTitle="No blocked tasks"
          page={1}
          pageSize={Math.max(data.blocked?.length ?? 1, 1)}
          totalItems={data.blocked?.length ?? 0}
          onPageChange={() => undefined}
        />
      </div>
      <EntityTable
        columns={[
          {
            key: "name",
            header: "Overdue",
            cell: (row) => (
              <Link className="underline-offset-4 hover:underline" to={`/projects/${row.projectId}`}>
                {row.name}
              </Link>
            ),
          },
          { key: "due", header: "Due", cell: (row) => row.dueDate ?? "—" },
        ]}
        rows={data.overdue ?? []}
        rowKey={(row) => row.id}
        emptyTitle="No overdue tasks"
        page={1}
        pageSize={Math.max(data.overdue?.length ?? 1, 1)}
        totalItems={data.overdue?.length ?? 0}
        onPageChange={() => undefined}
      />
      <EntityTable
        columns={[
          {
            key: "name",
            header: "Unassigned",
            cell: (row) => (
              <Link className="underline-offset-4 hover:underline" to={`/projects/${row.projectId}`}>
                {row.name}
              </Link>
            ),
          },
          {
            key: "priority",
            header: "Priority",
            cell: (row) => <PriorityBadge value={row.priority} />,
          },
        ]}
        rows={data.unassigned ?? []}
        rowKey={(row) => row.id}
        emptyTitle="No unassigned tasks"
        page={1}
        pageSize={Math.max(data.unassigned?.length ?? 1, 1)}
        totalItems={data.unassigned?.length ?? 0}
        onPageChange={() => undefined}
      />
      <EntityTable
        columns={[
          { key: "name", header: "Team capacity", cell: (row) => row.name },
          {
            key: "util",
            header: "%",
            cell: (row) => <CapacityCell percent={row.utilizationPercent} band={row.band} />,
          },
        ]}
        rows={data.teamCapacity ?? []}
        rowKey={(row) => row.userId}
        emptyTitle="No team capacity data"
        page={1}
        pageSize={Math.max(data.teamCapacity?.length ?? 1, 1)}
        totalItems={data.teamCapacity?.length ?? 0}
        onPageChange={() => undefined}
      />
      <EntityTable
        columns={[
          { key: "action", header: "Recent activity", cell: (row) => row.action },
          { key: "entity", header: "Entity", cell: (row) => row.entityName || row.entityType },
          {
            key: "when",
            header: "When",
            className: "px-3 py-2 font-mono tabular-nums",
            cell: (row) => new Date(row.createdAt).toLocaleString(),
          },
        ]}
        rows={data.recentActivity ?? []}
        rowKey={(row) => row.id}
        emptyTitle="No recent activity"
        page={1}
        pageSize={Math.max(data.recentActivity?.length ?? 1, 1)}
        totalItems={data.recentActivity?.length ?? 0}
        onPageChange={() => undefined}
        renderExpanded={(row) => <JsonDiff before={row.before} after={row.after} />}
      />
    </>
  )
}

function MemberView({ data }: { data: MemberDashboard }) {
  return (
    <>
      <EntityTable
        columns={[
          {
            key: "name",
            header: "Assigned tasks",
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
          { key: "due", header: "Due", cell: (row) => row.dueDate ?? "—" },
        ]}
        rows={data.tasks ?? []}
        rowKey={(row) => row.id}
        emptyTitle="No assigned tasks"
        page={1}
        pageSize={Math.max(data.tasks?.length ?? 1, 1)}
        totalItems={data.tasks?.length ?? 0}
        onPageChange={() => undefined}
      />
      <ChartCard
        title="This week’s load"
        description="Planned utilization"
        data={(data.weekLoad?.weekSeries ?? []).map((bucket) => ({
          period: bucket.periodKey,
          value: Math.round(bucket.utilizationPercent),
        }))}
      />
      {data.notifications?.length ? (
        <div className="flex flex-col gap-2">
          <h2 className="text-sm font-medium">Latest notifications</h2>
          {data.notifications.map((item) => (
            <NotificationItem key={item.id} item={item} />
          ))}
        </div>
      ) : (
        <EmptyState title="No notifications yet" />
      )}
    </>
  )
}

export function DashboardPage() {
  const role = useSession((state) => state.role)
  const kind = dashboardKind(role)
  const fetchDashboard = Boolean(role) && usesDashboardApi(kind)
  const dash = useQuery({
    queryKey: queryKeys.dashboard(kind),
    queryFn: () => getDashboard(kind),
    enabled: fetchDashboard,
  })

  const error =
    dash.error instanceof ApiError ? dash.error.message : dash.error ? "Could not load dashboard" : null

  const description =
    kind === "resource-manager"
      ? "Overload looks 14 days ahead. Availability uses the 4-week window from Resources."
      : "Summary for your role. Late projects and utilization come from the API."

  return (
    <div className="flex flex-col gap-4">
      <PageHeader
        title="Dashboard"
        description={description}
        actions={
          kind === "resource-manager" ? (
            <Link className={buttonVariants({ variant: "outline" })} to="/resources">
              View resources
            </Link>
          ) : null
        }
      />
      {error ? (
        <Alert variant="destructive">
          <AlertTitle>Could not load dashboard</AlertTitle>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      ) : null}
      {!role || (fetchDashboard && dash.isLoading) ? <DashboardSkeleton /> : null}
      {kind === "executive" && dash.data ? (
        <ExecutiveView data={dash.data as ExecutiveDashboard} />
      ) : null}
      {kind === "project-manager" && dash.data ? (
        <ProjectManagerView data={dash.data as ProjectManagerDashboard} />
      ) : null}
      {kind === "resource-manager" && role ? <ResourceManagerView /> : null}
      {kind === "member" && dash.data ? <MemberView data={dash.data as MemberDashboard} /> : null}
    </div>
  )
}
