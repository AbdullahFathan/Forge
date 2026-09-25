import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { useState } from "react"
import { Link } from "react-router"
import { toast } from "sonner"

import { EntityTable } from "@/components/common/entity-table"
import { FilterBar } from "@/components/common/filter-bar"
import { PageHeader } from "@/components/common/page-header"
import { PriorityBadge } from "@/components/common/priority-badge"
import { StatusBadge } from "@/components/common/status-badge"
import { projectStatusMap } from "@/components/common/status-map"
import { Button } from "@/components/ui/button"
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog"
import { listDepartments } from "@/features/departments/api/departments"
import { archiveProject, createProject, listProjects, patchProject } from "@/features/projects/api/projects"
import { ProjectFormDialog } from "@/features/projects/components/project-form-dialog"
import { toProjectPayload, type ProjectFormValues } from "@/features/projects/schema"
import { can } from "@/lib/auth"
import { canArchiveProject } from "@/lib/lifecycle"
import { compactParams } from "@/lib/search-params"
import { DEFAULT_PAGE_SIZE, PERMISSIONS } from "@/lib/constants"
import { useSession } from "@/lib/session"
import { ARCHIVE_PROJECT_COPY, type ProjectPublic } from "@/types/project"
import { ApiError } from "@/types/api"
import { queryKeys } from "@/services/query/query-keys"

const ALL = "all"

export function ProjectsPage() {
  const queryClient = useQueryClient()
  const user = useSession((state) => state.user)
  const permissions = useSession((state) => state.permissions)
  const canCreate = can(permissions, PERMISSIONS.projectCreate)
  const canDelete = can(permissions, PERMISSIONS.projectDelete)
  const canPickDepartment = can(permissions, PERMISSIONS.departmentManage)

  const [page, setPage] = useState(1)
  const [status, setStatus] = useState(ALL)
  const [departmentId, setDepartmentId] = useState(ALL)
  const [open, setOpen] = useState(false)
  const [editing, setEditing] = useState<ProjectPublic | null>(null)
  const [archiving, setArchiving] = useState<ProjectPublic | null>(null)

  const departments = useQuery({
    queryKey: queryKeys.departments({ page: 1, pageSize: 100, picker: true }),
    queryFn: () => listDepartments(1, 100),
    enabled: canPickDepartment,
  })

  const params = compactParams({
    page,
    pageSize: DEFAULT_PAGE_SIZE,
    status: status === ALL ? undefined : status,
    departmentId: departmentId === ALL ? undefined : departmentId,
  })

  const list = useQuery({
    queryKey: queryKeys.projects(params),
    queryFn: () =>
      listProjects({
        page,
        pageSize: DEFAULT_PAGE_SIZE,
        ...(status !== ALL ? { status } : {}),
        ...(departmentId !== ALL ? { departmentId } : {}),
      }),
  })

  const save = useMutation({
    mutationFn: (values: ProjectFormValues) => {
      if (!user) throw new Error("Not signed in")
      const payload = toProjectPayload(values, user.id, Boolean(editing))
      return editing ? patchProject(editing.id, payload) : createProject(payload as Parameters<typeof createProject>[0])
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["projects"] })
      toast.success(editing ? "Project updated" : "Project created")
      setOpen(false)
      setEditing(null)
    },
    onError: (error) => {
      toast.error(error instanceof ApiError ? error.message : "Could not save project")
    },
  })

  const archive = useMutation({
    mutationFn: (id: string) => archiveProject(id),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["projects"] })
      toast.success("Project archived")
      setArchiving(null)
    },
    onError: (error) => {
      toast.error(error instanceof ApiError ? error.message : "Could not archive project")
    },
  })

  const error =
    list.error instanceof ApiError
      ? list.error.status === 403
        ? "You cannot list projects."
        : list.error.message
      : list.error
        ? "Could not load projects"
        : null

  const createButton = canCreate ? (
    <Button
      onClick={() => {
        setEditing(null)
        save.reset()
        setOpen(true)
      }}
    >
      New project
    </Button>
  ) : null

  return (
    <div className="flex flex-col gap-4">
      <PageHeader
        title="Projects"
        description="Work scoped to your membership unless you can read all projects."
        actions={createButton}
      />
      <FilterBar>
        <Select
          value={status}
          onValueChange={(value) => {
            setStatus(value ?? ALL)
            setPage(1)
          }}
        >
          <SelectTrigger className="w-44" aria-label="Filter by status">
            <SelectValue placeholder="All statuses" />
          </SelectTrigger>
          <SelectContent>
            <SelectGroup>
              <SelectItem value={ALL}>All statuses</SelectItem>
              {Object.entries(projectStatusMap).map(([value, visual]) => (
                <SelectItem key={value} value={value}>
                  {visual.label}
                </SelectItem>
              ))}
            </SelectGroup>
          </SelectContent>
        </Select>
        {canPickDepartment ? (
          <Select
            items={{
              [ALL]: "All departments",
              ...Object.fromEntries((departments.data?.items ?? []).map((department) => [department.id, department.name])),
            }}
            value={departmentId}
            onValueChange={(value) => {
              setDepartmentId(value ?? ALL)
              setPage(1)
            }}
          >
            <SelectTrigger className="w-52" aria-label="Filter by department">
              <SelectValue placeholder="All departments" />
            </SelectTrigger>
            <SelectContent>
              <SelectGroup>
                <SelectItem value={ALL}>All departments</SelectItem>
                {(departments.data?.items ?? []).map((department) => (
                  <SelectItem key={department.id} value={department.id}>
                    {department.name}
                  </SelectItem>
                ))}
              </SelectGroup>
            </SelectContent>
          </Select>
        ) : null}
      </FilterBar>
      <EntityTable
        columns={[
          {
            key: "name",
            header: "Name",
            cell: (row) => (
              <Link className="font-medium underline-offset-4 hover:underline" to={`/projects/${row.id}`}>
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
            key: "priority",
            header: "Priority",
            cell: (row) => <PriorityBadge value={row.priority} />,
          },
          {
            key: "progress",
            header: "Progress",
            className: "px-3 py-2 font-mono tabular-nums",
            cell: (row) => `${Math.round(row.completionPercent)}%`,
          },
          {
            key: "dates",
            header: "Dates",
            cell: (row) => (
              <span className="tabular-nums">
                {row.startDate} → {row.targetEndDate}
              </span>
            ),
          },
          {
            key: "owner",
            header: "Owner",
            cell: (row) => row.ownerName ?? "—",
          },
          {
            key: "actions",
            header: "",
            cell: (row) => (
              <div className="flex justify-end gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => {
                    setEditing(row)
                    save.reset()
                    setOpen(true)
                  }}
                >
                  Edit
                </Button>
                {canDelete && canArchiveProject(row.status) ? (
                  <Button variant="outline" size="sm" onClick={() => setArchiving(row)}>
                    Archive
                  </Button>
                ) : null}
              </div>
            ),
          },
        ]}
        rows={list.data?.items ?? []}
        rowKey={(row) => row.id}
        isLoading={list.isLoading}
        error={error}
        emptyTitle="No projects yet"
        emptyDescription="Create a project if you have permission, or wait to be added as a member."
        emptyAction={createButton}
        page={page}
        pageSize={list.data?.pageSize ?? DEFAULT_PAGE_SIZE}
        totalItems={list.data?.totalItems ?? 0}
        onPageChange={setPage}
      />
      <ProjectFormDialog
        open={open}
        project={editing}
        departments={departments.data?.items ?? []}
        canPickDepartment={canPickDepartment}
        pending={save.isPending}
        error={save.error}
        onOpenChange={setOpen}
        onSubmit={(values) => save.mutate(values)}
      />
      <AlertDialog open={Boolean(archiving)} onOpenChange={(next) => !next && setArchiving(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{ARCHIVE_PROJECT_COPY}</AlertDialogTitle>
            <AlertDialogDescription>
              {archiving
                ? `Archive “${archiving.name}”? Historical rows stay. This is not a hard delete.`
                : ""}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              variant="destructive"
              disabled={archive.isPending}
              onClick={() => archiving && archive.mutate(archiving.id)}
            >
              Archive
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
