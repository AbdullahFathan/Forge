import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { LayoutGridIcon, ListIcon } from "lucide-react"
import { useMemo, useState } from "react"
import { toast } from "sonner"

import { EntityTable } from "@/components/common/entity-table"
import { FilterBar } from "@/components/common/filter-bar"
import { PriorityBadge } from "@/components/common/priority-badge"
import { StatusBadge } from "@/components/common/status-badge"
import { UserPile } from "@/components/common/user-pile"
import { taskStatusMap, priorityMap } from "@/components/common/status-map"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group"
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
import { createTask, deleteTask, listProjectTasks, patchTask } from "@/features/tasks/api/tasks"
import { AddSubtaskButton } from "@/features/tasks/components/add-subtask-button"
import { KanbanBoard } from "@/features/tasks/components/kanban-board"
import { TaskFormDialog } from "@/features/tasks/components/task-form-dialog"
import { TaskSheet } from "@/features/tasks/components/task-sheet"
import { collectTaskWarnings, flattenTasks } from "@/features/tasks/flatten"
import { toTaskCreate, toTaskPatch, type TaskFormValues } from "@/features/tasks/schema"
import { useMediaQuery } from "@/hooks/use-media-query"
import { can } from "@/lib/auth"
import {
  canAssigneeOnlyPatch,
  canManageProject,
  canMutateTaskFields,
  isProjectViewer,
  projectRoleFor,
} from "@/lib/project-access"
import { DEFAULT_PAGE_SIZE, PERMISSIONS } from "@/lib/constants"
import { TASK_STATUSES } from "@/lib/lifecycle"
import { compactParams } from "@/lib/search-params"
import { useSession } from "@/lib/session"
import { ApiError } from "@/types/api"
import type { ProjectMember, ProjectPublic } from "@/types/project"
import type { TaskPublic } from "@/types/task"
import { queryKeys } from "@/services/query/query-keys"

const ALL = "all"

export function ProjectTasksTab({
  project,
  members,
}: {
  project: ProjectPublic
  members: ProjectMember[]
}) {
  const queryClient = useQueryClient()
  const user = useSession((state) => state.user)
  const role = useSession((state) => state.role)
  const permissions = useSession((state) => state.permissions)
  const reducedMotion = useMediaQuery("(prefers-reduced-motion: reduce)")
  const manage = canManageProject(project, members, user, role)
  const viewer = isProjectViewer(projectRoleFor(members, user?.id))
  const canTaskManage = can(permissions, PERMISSIONS.taskManage) && !viewer
  const canCreate = canTaskManage && manage

  const [view, setView] = useState<"list" | "board">("list")
  const [page, setPage] = useState(1)
  const [status, setStatus] = useState(ALL)
  const [assignee, setAssignee] = useState(ALL)
  const [priority, setPriority] = useState(ALL)
  const [dueFrom, setDueFrom] = useState("")
  const [dueTo, setDueTo] = useState("")
  const [label, setLabel] = useState("")
  const [open, setOpen] = useState(false)
  const [editing, setEditing] = useState<TaskPublic | null>(null)
  const [parent, setParent] = useState<TaskPublic | null>(null)
  const [removing, setRemoving] = useState<TaskPublic | null>(null)
  const [selected, setSelected] = useState<TaskPublic | null>(null)

  const params = compactParams({
    page,
    pageSize: DEFAULT_PAGE_SIZE,
    status: status === ALL ? undefined : status,
    assignee: assignee === ALL ? undefined : assignee,
    priority: priority === ALL ? undefined : priority,
    dueFrom: dueFrom || undefined,
    dueTo: dueTo || undefined,
    label: label.trim() || undefined,
    include: "subtasks",
  })

  const list = useQuery({
    queryKey: queryKeys.projectTasks(project.id, params),
    queryFn: () =>
      listProjectTasks(project.id, {
        page,
        pageSize: DEFAULT_PAGE_SIZE,
        include: "subtasks",
        ...(status !== ALL ? { status } : {}),
        ...(assignee !== ALL ? { assignee } : {}),
        ...(priority !== ALL ? { priority } : {}),
        ...(dueFrom ? { dueFrom } : {}),
        ...(dueTo ? { dueTo } : {}),
        ...(label.trim() ? { label: label.trim() } : {}),
      }),
  })

  const rows = useMemo(() => flattenTasks(list.data?.items ?? []), [list.data?.items])
  const warnings = useMemo(() => collectTaskWarnings(list.data?.items ?? []), [list.data?.items])

  const save = useMutation({
    mutationFn: (values: TaskFormValues) => {
      if (editing) return patchTask(editing.id, toTaskPatch(values))
      return createTask(project.id, toTaskCreate(values, parent?.id ?? null))
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["projects", project.id] })
      toast.success(editing ? "Task updated" : "Task created")
      setOpen(false)
      setEditing(null)
      setParent(null)
    },
    onError: (error) => {
      toast.error(error instanceof ApiError ? error.message : "Could not save task")
    },
  })

  const remove = useMutation({
    mutationFn: (id: string) => deleteTask(id),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["projects", project.id] })
      toast.success("Task removed")
      setRemoving(null)
    },
    onError: (error) => {
      toast.error(error instanceof ApiError ? error.message : "Could not remove task")
    },
  })

  const move = useMutation({
    mutationFn: ({ id, status, position }: { id: string; status: string; position: number }) =>
      patchTask(id, { status, position }),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["projects", project.id, "tasks"] })
    },
    onError: (error) => {
      toast.error(error instanceof ApiError ? error.message : "Could not update status")
    },
  })

  const error =
    list.error instanceof ApiError ? list.error.message : list.error ? "Could not load tasks" : null

  function openCreate(parentTask?: TaskPublic) {
    setEditing(null)
    setParent(parentTask ?? null)
    save.reset()
    setOpen(true)
  }

  function canChangeStatus(task: TaskPublic) {
    return canMutateTaskFields(task, project, members, user, role, permissions)
  }

  function canFullEdit(task: TaskPublic) {
    return canChangeStatus(task) && !canAssigneeOnlyPatch(task, project, members, user, role)
  }

  const allForBoard = rows.map((row) => row.task)

  return (
    <div className="flex flex-col gap-3">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <ToggleGroup
          value={[view]}
          onValueChange={(value) => {
            const next = value[0]
            if (next === "list" || next === "board") setView(next)
          }}
          spacing={0}
          className="rounded-lg border"
        >
          <ToggleGroupItem value="list" aria-label="List view">
            <ListIcon data-icon="inline-start" />
            List
          </ToggleGroupItem>
          <ToggleGroupItem value="board" aria-label="Board view">
            <LayoutGridIcon data-icon="inline-start" />
            Board
          </ToggleGroupItem>
        </ToggleGroup>
        {canCreate ? (
          <Button size="sm" onClick={() => openCreate()}>
            New task
          </Button>
        ) : null}
      </div>
      <FilterBar>
        <Select
          value={status}
          onValueChange={(value) => {
            setStatus(value ?? ALL)
            setPage(1)
          }}
        >
          <SelectTrigger className="w-40" aria-label="Filter by status">
            <SelectValue placeholder="Status" />
          </SelectTrigger>
          <SelectContent>
            <SelectGroup>
              <SelectItem value={ALL}>All statuses</SelectItem>
              {TASK_STATUSES.map((value) => (
                <SelectItem key={value} value={value}>
                  {taskStatusMap[value]?.label ?? value}
                </SelectItem>
              ))}
            </SelectGroup>
          </SelectContent>
        </Select>
        <Select
          value={assignee}
          onValueChange={(value) => {
            setAssignee(value ?? ALL)
            setPage(1)
          }}
        >
          <SelectTrigger className="w-44" aria-label="Filter by assignee">
            <SelectValue placeholder="Assignee" />
          </SelectTrigger>
          <SelectContent>
            <SelectGroup>
              <SelectItem value={ALL}>All assignees</SelectItem>
              {members.map((member) => (
                <SelectItem key={member.userId} value={member.userId}>
                  {member.name}
                </SelectItem>
              ))}
            </SelectGroup>
          </SelectContent>
        </Select>
        <Select
          value={priority}
          onValueChange={(value) => {
            setPriority(value ?? ALL)
            setPage(1)
          }}
        >
          <SelectTrigger className="w-36" aria-label="Filter by priority">
            <SelectValue placeholder="Priority" />
          </SelectTrigger>
          <SelectContent>
            <SelectGroup>
              <SelectItem value={ALL}>All priorities</SelectItem>
              {Object.entries(priorityMap).map(([value, visual]) => (
                <SelectItem key={value} value={value}>
                  {visual.label}
                </SelectItem>
              ))}
            </SelectGroup>
          </SelectContent>
        </Select>
        <Input
          type="date"
          aria-label="Due from"
          value={dueFrom}
          onChange={(event) => {
            setDueFrom(event.target.value)
            setPage(1)
          }}
          className="w-40"
        />
        <Input
          type="date"
          aria-label="Due to"
          value={dueTo}
          onChange={(event) => {
            setDueTo(event.target.value)
            setPage(1)
          }}
          className="w-40"
        />
        <Input
          aria-label="Label"
          placeholder="Label"
          value={label}
          onChange={(event) => {
            setLabel(event.target.value)
            setPage(1)
          }}
          className="w-36"
        />
      </FilterBar>
      {warnings.length > 0 ? (
        <Alert>
          <AlertTitle>Finish-to-start warnings</AlertTitle>
          <AlertDescription>
            {warnings.length} task{warnings.length === 1 ? "" : "s"} have predecessors that are not Done.
          </AlertDescription>
        </Alert>
      ) : null}
      {view === "list" ? (
        <EntityTable
          columns={[
            {
              key: "name",
              header: "Task",
              cell: (row) => (
                <button
                  type="button"
                  className="text-left font-medium underline-offset-4 hover:underline"
                  style={{ paddingInlineStart: row.depth * 16 }}
                  onClick={() => setSelected(row.task)}
                >
                  {row.task.name}
                </button>
              ),
            },
            {
              key: "status",
              header: "Status",
              cell: (row) => <StatusBadge kind="task" value={row.task.status} />,
            },
            {
              key: "priority",
              header: "Priority",
              cell: (row) => <PriorityBadge value={row.task.priority} />,
            },
            {
              key: "assignees",
              header: "Assignees",
              cell: (row) => <UserPile people={row.task.assignees} />,
            },
            {
              key: "due",
              header: "Due",
              className: "px-3 py-2 tabular-nums",
              cell: (row) => row.task.dueDate ?? "—",
            },
            {
              key: "actions",
              header: "",
              cell: (row) => (
                <div className="flex justify-end gap-1">
                  {canCreate ? (
                    <AddSubtaskButton parentTaskId={row.task.parentTaskId} onClick={() => openCreate(row.task)} />
                  ) : null}
                  {canFullEdit(row.task) ? (
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => {
                        setParent(null)
                        setEditing(row.task)
                        save.reset()
                        setOpen(true)
                      }}
                    >
                      Edit
                    </Button>
                  ) : null}
                  {canFullEdit(row.task) ? (
                    <Button variant="outline" size="sm" onClick={() => setRemoving(row.task)}>
                      Remove
                    </Button>
                  ) : null}
                </div>
              ),
            },
          ]}
          rows={rows}
          rowKey={(row) => row.task.id}
          isLoading={list.isLoading}
          error={error}
          emptyTitle="No tasks"
          emptyDescription="Create a task or adjust filters."
          emptyAction={
            canCreate ? (
              <Button size="sm" onClick={() => openCreate()}>
                New task
              </Button>
            ) : undefined
          }
          page={page}
          pageSize={list.data?.pageSize ?? DEFAULT_PAGE_SIZE}
          totalItems={list.data?.totalItems ?? 0}
          onPageChange={setPage}
        />
      ) : (
        <KanbanBoard
          tasks={allForBoard}
          reducedMotion={reducedMotion}
          canChangeStatus={canChangeStatus}
          onOpen={setSelected}
          onMove={(id, nextStatus, position) => move.mutate({ id, status: nextStatus, position })}
        />
      )}
      <TaskFormDialog
        open={open}
        task={editing}
        parentName={parent?.name}
        members={members}
        pending={save.isPending}
        error={save.error}
        onOpenChange={setOpen}
        onSubmit={(values) => save.mutate(values)}
      />
      <TaskSheet
        task={selected}
        allTasks={list.data?.items ?? []}
        canComment={Boolean(selected && canChangeStatus(selected))}
        canManageDeps={Boolean(selected && canFullEdit(selected))}
        onOpenChange={(next) => {
          if (!next) setSelected(null)
        }}
      />
      <AlertDialog open={Boolean(removing)} onOpenChange={(next) => !next && setRemoving(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Remove task</AlertDialogTitle>
            <AlertDialogDescription>
              {removing ? `Remove “${removing.name}”? This is a soft delete.` : ""}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              variant="destructive"
              disabled={remove.isPending}
              onClick={() => removing && remove.mutate(removing.id)}
            >
              Remove
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
