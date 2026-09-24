import { useQuery } from "@tanstack/react-query"
import { useEffect } from "react"
import { useParams } from "react-router"

import { ForbiddenPage } from "@/components/common/forbidden-page"
import { Loader } from "@/components/common/loader"
import { PageHeader } from "@/components/common/page-header"
import { PriorityBadge } from "@/components/common/priority-badge"
import { StatusBadge } from "@/components/common/status-badge"
import { EmptyState } from "@/components/common/empty-state"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { listProjectMembers } from "@/features/projects/api/members"
import { getProject } from "@/features/projects/api/projects"
import { ProjectActivityTab } from "@/features/projects/components/project-activity-tab"
import { ProjectMembersTab } from "@/features/projects/components/project-members-tab"
import { ProjectOverviewTimeline } from "@/features/projects/components/project-overview-timeline"
import { ProjectTasksTab } from "@/features/tasks/components/project-tasks-tab"
import { usePageCrumb } from "@/lib/breadcrumb"
import { canManageProject } from "@/lib/project-access"
import { useSession } from "@/lib/session"
import { ApiError } from "@/types/api"
import { queryKeys } from "@/services/query/query-keys"
import { taskStatusMap } from "@/components/common/status-map"

export function ProjectDetailPage() {
  const { id = "" } = useParams()
  const user = useSession((state) => state.user)
  const role = useSession((state) => state.role)
  const setDetailLabel = usePageCrumb((state) => state.setDetailLabel)

  const project = useQuery({
    queryKey: queryKeys.project(id),
    queryFn: () => getProject(id),
    enabled: Boolean(id),
  })

  const members = useQuery({
    queryKey: queryKeys.projectMembers(id, { page: 1, pageSize: 100, picker: true }),
    queryFn: () => listProjectMembers(id, 1, 100),
    enabled: Boolean(id) && project.isSuccess,
  })

  useEffect(() => {
    setDetailLabel(project.data?.name ?? null)
    return () => setDetailLabel(null)
  }, [project.data?.name, setDetailLabel])

  if (project.isLoading) return <Loader />
  if (project.error instanceof ApiError && project.error.status === 403) return <ForbiddenPage />
  if (project.error instanceof ApiError && project.error.status === 404) {
    return <EmptyState title="Project not found" description="It may have been archived or the id is wrong." />
  }
  if (project.error || !project.data) {
    return (
      <EmptyState
        title="Could not load project"
        description={project.error instanceof ApiError ? project.error.message : "Try again later."}
      />
    )
  }

  const summary = project.data
  const counts = summary.taskCounts
  const manage = canManageProject(summary, members.data?.items, user, role)

  return (
    <div className="flex flex-col gap-4">
      <PageHeader
        title={summary.name}
        description={summary.description || `${summary.startDate} → ${summary.targetEndDate}`}
        actions={
          <div className="flex items-center gap-2">
            <StatusBadge kind="project" value={summary.status} />
            <PriorityBadge value={summary.priority} />
          </div>
        }
      />
      <Tabs defaultValue="overview">
        <TabsList>
          <TabsTrigger value="overview">Overview</TabsTrigger>
          <TabsTrigger value="tasks">Tasks</TabsTrigger>
          <TabsTrigger value="members">Members</TabsTrigger>
          <TabsTrigger value="activity">Activity</TabsTrigger>
        </TabsList>
        <TabsContent value="overview" className="flex flex-col gap-3 pt-3">
          <div className="grid gap-3 sm:grid-cols-3">
            <Card size="sm">
              <CardHeader>
                <CardTitle>Progress</CardTitle>
              </CardHeader>
              <CardContent className="font-mono text-2xl tabular-nums">
                {Math.round(summary.completionPercent)}%
              </CardContent>
            </Card>
            <Card size="sm">
              <CardHeader>
                <CardTitle>Members</CardTitle>
              </CardHeader>
              <CardContent className="font-mono text-2xl tabular-nums">{summary.memberCount}</CardContent>
            </Card>
            <Card size="sm">
              <CardHeader>
                <CardTitle>Tasks</CardTitle>
              </CardHeader>
              <CardContent className="font-mono text-2xl tabular-nums">{counts.total}</CardContent>
            </Card>
          </div>
          <ProjectOverviewTimeline
            startDate={summary.startDate}
            targetEndDate={summary.targetEndDate}
            completionPercent={summary.completionPercent}
            status={summary.status}
          />
          <Card>
            <CardHeader>
              <CardTitle>Status counts</CardTitle>
            </CardHeader>
            <CardContent className="flex flex-wrap gap-3">
              {Object.entries(taskStatusMap).map(([status, visual]) => (
                <p key={status} className="text-sm">
                  {visual.label}:{" "}
                  <span className="font-mono tabular-nums">
                    {counts[status as keyof typeof counts] ?? 0}
                  </span>
                </p>
              ))}
            </CardContent>
          </Card>
          <p className="text-sm text-muted-foreground">
            Owner {summary.ownerName ?? summary.ownerId}
            {summary.departmentName ? ` · ${summary.departmentName}` : ""}
          </p>
        </TabsContent>
        <TabsContent value="tasks" className="pt-3">
          <ProjectTasksTab project={summary} members={members.data?.items ?? []} />
        </TabsContent>
        <TabsContent value="members" className="pt-3">
          <ProjectMembersTab project={summary} canManage={manage} />
        </TabsContent>
        <TabsContent value="activity" className="pt-3">
          <ProjectActivityTab projectId={summary.id} />
        </TabsContent>
      </Tabs>
    </div>
  )
}
