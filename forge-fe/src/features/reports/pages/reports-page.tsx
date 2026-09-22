import { useQuery } from "@tanstack/react-query"
import { useState } from "react"
import { toast } from "sonner"

import { EntityTable } from "@/components/common/entity-table"
import { FilterBar } from "@/components/common/filter-bar"
import { PageHeader } from "@/components/common/page-header"
import { StatusBadge } from "@/components/common/status-badge"
import { CapacityCell } from "@/components/common/capacity-cell"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
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
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { listDepartments } from "@/features/departments/api/departments"
import { listProjects } from "@/features/projects/api/projects"
import {
  getProjectStatusReport,
  getTaskCompletionReport,
  getUtilizationReport,
} from "@/features/reports/api/reports"
import { can } from "@/lib/auth"
import { downloadBlob } from "@/lib/download"
import { compactParams } from "@/lib/search-params"
import { PERMISSIONS } from "@/lib/constants"
import { useSession } from "@/lib/session"
import { endpoints } from "@/services/api/endpoints"
import { queryKeys } from "@/services/query/query-keys"
import { ApiError } from "@/types/api"
import type { ReportFilters, ReportKind } from "@/types/report"

const ALL = "all"

const reportPaths: Record<ReportKind, string> = {
  "project-status": endpoints.reportsProjectStatus,
  "resource-utilization": endpoints.reportsUtilization,
  "task-completion": endpoints.reportsTaskCompletion,
}

export function ReportsPage() {
  const permissions = useSession((state) => state.permissions)
  const canPickDepartment = can(permissions, PERMISSIONS.departmentManage)
  const [kind, setKind] = useState<ReportKind>("project-status")
  const [from, setFrom] = useState("")
  const [to, setTo] = useState("")
  const [departmentId, setDepartmentId] = useState(ALL)
  const [projectId, setProjectId] = useState(ALL)
  const [groupBy, setGroupBy] = useState<"project" | "user">("project")
  const [downloading, setDownloading] = useState<"csv" | "pdf" | null>(null)

  const departments = useQuery({
    queryKey: queryKeys.departments({ page: 1, pageSize: 100, picker: true }),
    queryFn: () => listDepartments(1, 100),
    enabled: canPickDepartment,
  })

  const projects = useQuery({
    queryKey: queryKeys.projects({ page: 1, pageSize: 100, picker: true }),
    queryFn: () => listProjects({ page: 1, pageSize: 100 }),
  })

  const filterInput: ReportFilters = {
    from: from || undefined,
    to: to || undefined,
    departmentId: departmentId === ALL ? undefined : departmentId,
    projectId: projectId === ALL ? undefined : projectId,
    groupBy: kind === "task-completion" ? groupBy : undefined,
  }
  const filters = compactParams(filterInput)

  const projectStatus = useQuery({
    queryKey: queryKeys.reports("project-status", filters),
    queryFn: () => getProjectStatusReport(filterInput),
    enabled: kind === "project-status",
  })
  const utilization = useQuery({
    queryKey: queryKeys.reports("resource-utilization", filters),
    queryFn: () => getUtilizationReport(filterInput),
    enabled: kind === "resource-utilization",
  })
  const completion = useQuery({
    queryKey: queryKeys.reports("task-completion", filters),
    queryFn: () => getTaskCompletionReport(filterInput),
    enabled: kind === "task-completion",
  })

  async function download(format: "csv" | "pdf") {
    setDownloading(format)
    try {
      await downloadBlob(reportPaths[kind], { ...filters, format }, `${kind}.${format}`)
    } catch (error) {
      toast.error(error instanceof ApiError ? error.message : `Could not download ${format.toUpperCase()}`)
    } finally {
      setDownloading(null)
    }
  }

  const activeError =
    kind === "project-status"
      ? projectStatus.error
      : kind === "resource-utilization"
        ? utilization.error
        : completion.error
  const error =
    activeError instanceof ApiError
      ? activeError.message
      : activeError
        ? "Could not load report"
        : null

  return (
    <div className="flex flex-col gap-4">
      <PageHeader
        title="Reports"
        description="Preview matches the API. Downloads use the same filters — numbers are not recomputed in the browser."
        actions={
          <div className="flex gap-2">
            <Button type="button" variant="outline" disabled={Boolean(downloading)} onClick={() => void download("csv")}>
              {downloading === "csv" ? <Spinner data-icon="inline-start" /> : null}
              CSV
            </Button>
            <Button type="button" variant="outline" disabled={Boolean(downloading)} onClick={() => void download("pdf")}>
              {downloading === "pdf" ? <Spinner data-icon="inline-start" /> : null}
              PDF
            </Button>
          </div>
        }
      />
      <FilterBar>
        <div className="flex flex-col gap-1">
          <Label htmlFor="report-from">From</Label>
          <Input id="report-from" type="date" value={from} onChange={(event) => setFrom(event.target.value)} />
        </div>
        <div className="flex flex-col gap-1">
          <Label htmlFor="report-to">To</Label>
          <Input id="report-to" type="date" value={to} onChange={(event) => setTo(event.target.value)} />
        </div>
        {canPickDepartment ? (
          <div className="flex flex-col gap-1">
            <Label>Department</Label>
            <Select value={departmentId} onValueChange={(value) => setDepartmentId(value ?? ALL)}>
              <SelectTrigger className="w-48">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  <SelectItem value={ALL}>All departments</SelectItem>
                  {(departments.data?.items ?? []).map((dept) => (
                    <SelectItem key={dept.id} value={dept.id}>
                      {dept.name}
                    </SelectItem>
                  ))}
                </SelectGroup>
              </SelectContent>
            </Select>
          </div>
        ) : null}
        <div className="flex flex-col gap-1">
          <Label>Project</Label>
          <Select value={projectId} onValueChange={(value) => setProjectId(value ?? ALL)}>
            <SelectTrigger className="w-56">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectGroup>
                <SelectItem value={ALL}>All projects</SelectItem>
                {(projects.data?.items ?? []).map((project) => (
                  <SelectItem key={project.id} value={project.id}>
                    {project.name}
                  </SelectItem>
                ))}
              </SelectGroup>
            </SelectContent>
          </Select>
        </div>
        {kind === "task-completion" ? (
          <div className="flex flex-col gap-1">
            <Label>Group by</Label>
            <Select value={groupBy} onValueChange={(value) => setGroupBy((value as "project" | "user") ?? "project")}>
              <SelectTrigger className="w-40">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  <SelectItem value="project">Project</SelectItem>
                  <SelectItem value="user">User</SelectItem>
                </SelectGroup>
              </SelectContent>
            </Select>
          </div>
        ) : null}
      </FilterBar>
      <Tabs
        value={kind}
        onValueChange={(value) => setKind((value as ReportKind) ?? "project-status")}
      >
        <TabsList>
          <TabsTrigger value="project-status">Project status</TabsTrigger>
          <TabsTrigger value="resource-utilization">Utilization</TabsTrigger>
          <TabsTrigger value="task-completion">Task completion</TabsTrigger>
        </TabsList>
        <TabsContent value="project-status">
          <EntityTable
            columns={[
              { key: "name", header: "Project", cell: (row) => row.name },
              {
                key: "status",
                header: "Status",
                cell: (row) => <StatusBadge kind="project" value={row.status} />,
              },
              {
                key: "pct",
                header: "Completion",
                cell: (row) => <span className="tabular-nums">{Math.round(row.completionPercent)}%</span>,
              },
              { key: "end", header: "Deadline", cell: (row) => row.targetEndDate },
              { key: "owner", header: "Owner", cell: (row) => row.ownerName ?? row.ownerId },
            ]}
            rows={projectStatus.data ?? []}
            rowKey={(row) => row.id}
            isLoading={projectStatus.isLoading}
            error={kind === "project-status" ? error : null}
            emptyTitle="No project rows"
            page={1}
            pageSize={Math.max(projectStatus.data?.length ?? 1, 1)}
            totalItems={projectStatus.data?.length ?? 0}
            onPageChange={() => undefined}
          />
        </TabsContent>
        <TabsContent value="resource-utilization">
          <EntityTable
            columns={[
              { key: "name", header: "Person", cell: (row) => row.name },
              {
                key: "util",
                header: "Utilization",
                cell: (row) => <CapacityCell percent={row.utilizationPercent} band={row.band} />,
              },
            ]}
            rows={utilization.data ?? []}
            rowKey={(row) => row.userId}
            isLoading={utilization.isLoading}
            error={kind === "resource-utilization" ? error : null}
            emptyTitle="No utilization rows"
            page={1}
            pageSize={Math.max(utilization.data?.length ?? 1, 1)}
            totalItems={utilization.data?.length ?? 0}
            onPageChange={() => undefined}
          />
        </TabsContent>
        <TabsContent value="task-completion">
          <EntityTable
            columns={[
              { key: "name", header: "Group", cell: (row) => row.name },
              { key: "done", header: "Done", cell: (row) => <span className="tabular-nums">{row.done}</span> },
              { key: "total", header: "Total", cell: (row) => <span className="tabular-nums">{row.total}</span> },
              {
                key: "pct",
                header: "Percent",
                cell: (row) => <span className="tabular-nums">{Math.round(row.percent)}%</span>,
              },
            ]}
            rows={completion.data ?? []}
            rowKey={(row) => row.id}
            isLoading={completion.isLoading}
            error={kind === "task-completion" ? error : null}
            emptyTitle="No completion rows"
            page={1}
            pageSize={Math.max(completion.data?.length ?? 1, 1)}
            totalItems={completion.data?.length ?? 0}
            onPageChange={() => undefined}
          />
        </TabsContent>
      </Tabs>
    </div>
  )
}
