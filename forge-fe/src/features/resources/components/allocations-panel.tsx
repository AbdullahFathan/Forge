import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { useState } from "react"
import { toast } from "sonner"

import { EntityTable } from "@/components/common/entity-table"
import { FilterBar } from "@/components/common/filter-bar"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
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
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { listProjects } from "@/features/projects/api/projects"
import {
  createAllocation,
  deleteAllocation,
  listAllocations,
  patchAllocation,
} from "@/features/resources/api/allocations"
import { AllocationFormDialog } from "@/features/resources/components/allocation-form-dialog"
import { OverAllocationBanner } from "@/features/resources/components/over-allocation-banner"
import {
  toAllocationCreate,
  toAllocationPatch,
  type AllocationFormValues,
} from "@/features/resources/schema"
import { compactParams } from "@/lib/search-params"
import { DEFAULT_PAGE_SIZE } from "@/lib/constants"
import { isOverAllocated } from "@/lib/capacity-band"
import { queryKeys } from "@/services/query/query-keys"
import { ApiError } from "@/types/api"
import type { AllocationPublic } from "@/types/resource"
import type { PublicUser } from "@/types/user"

const ALL = "all"

export function AllocationsPanel({
  users,
  canPickUsers,
  formDefaults,
  onFormDefaultsConsumed,
}: {
  users: PublicUser[]
  canPickUsers: boolean
  formDefaults?: Partial<AllocationFormValues> | null
  onFormDefaultsConsumed?: () => void
}) {
  const queryClient = useQueryClient()
  const [page, setPage] = useState(1)
  const [userId, setUserId] = useState(ALL)
  const [projectId, setProjectId] = useState(ALL)
  const [from, setFrom] = useState("")
  const [to, setTo] = useState("")
  const [open, setOpen] = useState(Boolean(formDefaults))
  const [editing, setEditing] = useState<AllocationPublic | null>(null)
  const [removing, setRemoving] = useState<AllocationPublic | null>(null)
  const [warning, setWarning] = useState<AllocationPublic | null>(null)

  const projects = useQuery({
    queryKey: queryKeys.projects({ page: 1, pageSize: 100, picker: true }),
    queryFn: () => listProjects({ page: 1, pageSize: 100 }),
  })

  const params = compactParams({
    page,
    pageSize: DEFAULT_PAGE_SIZE,
    userId: userId === ALL ? undefined : userId,
    projectId: projectId === ALL ? undefined : projectId,
    from: from || undefined,
    to: to || undefined,
  })

  const list = useQuery({
    queryKey: queryKeys.allocations(params),
    queryFn: () =>
      listAllocations({
        page,
        pageSize: DEFAULT_PAGE_SIZE,
        ...(userId !== ALL ? { userId } : {}),
        ...(projectId !== ALL ? { projectId } : {}),
        ...(from ? { from } : {}),
        ...(to ? { to } : {}),
      }),
  })

  const save = useMutation({
    mutationFn: (values: AllocationFormValues) =>
      editing
        ? patchAllocation(editing.id, toAllocationPatch(values))
        : createAllocation(toAllocationCreate(values)),
    onSuccess: async (row) => {
      await queryClient.invalidateQueries({ queryKey: ["resources"] })
      setOpen(false)
      setEditing(null)
      onFormDefaultsConsumed?.()
      if (isOverAllocated(row.overAllocated, row.warnings)) {
        setWarning(row)
        return
      }
      setWarning(null)
      toast.success(editing ? "Allocation updated" : "Allocation saved")
    },
  })

  const remove = useMutation({
    mutationFn: (id: string) => deleteAllocation(id),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["resources"] })
      toast.success("Allocation removed")
      setRemoving(null)
    },
    onError: (error) => {
      toast.error(error instanceof ApiError ? error.message : "Could not remove allocation")
    },
  })

  return (
    <div className="flex flex-col gap-3">
      {warning ? (
        <OverAllocationBanner overAllocated={warning.overAllocated} warnings={warning.warnings} />
      ) : null}
      <FilterBar>
        {canPickUsers ? (
          <Select
            value={userId}
            onValueChange={(value) => {
              setUserId(value ?? ALL)
              setPage(1)
            }}
          >
            <SelectTrigger className="w-48">
              <SelectValue placeholder="All people" />
            </SelectTrigger>
            <SelectContent>
              <SelectGroup>
                <SelectItem value={ALL}>All people</SelectItem>
                {users.map((user) => (
                  <SelectItem key={user.id} value={user.id}>
                    {user.name}
                  </SelectItem>
                ))}
              </SelectGroup>
            </SelectContent>
          </Select>
        ) : null}
        <Select
          value={projectId}
          onValueChange={(value) => {
            setProjectId(value ?? ALL)
            setPage(1)
          }}
        >
          <SelectTrigger className="w-48">
            <SelectValue placeholder="All projects" />
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
        <div className="flex flex-col gap-1">
          <Label htmlFor="alloc-from">From</Label>
          <Input
            id="alloc-from"
            type="date"
            value={from}
            onChange={(event) => {
              setFrom(event.target.value)
              setPage(1)
            }}
          />
        </div>
        <div className="flex flex-col gap-1">
          <Label htmlFor="alloc-to">To</Label>
          <Input
            id="alloc-to"
            type="date"
            value={to}
            onChange={(event) => {
              setTo(event.target.value)
              setPage(1)
            }}
          />
        </div>
        <Button
          onClick={() => {
            setEditing(null)
            setOpen(true)
          }}
        >
          New allocation
        </Button>
      </FilterBar>
      <EntityTable
        columns={[
          { key: "user", header: "Person", cell: (row) => row.user.name },
          { key: "project", header: "Project", cell: (row) => row.projectName ?? row.projectId },
          {
            key: "percent",
            header: "%",
            cell: (row) => <span className="tabular-nums">{row.allocationPercent}%</span>,
          },
          {
            key: "dates",
            header: "Period",
            cell: (row) => `${row.startDate} → ${row.endDate}`,
          },
          { key: "role", header: "Role", cell: (row) => row.role },
          {
            key: "warn",
            header: "Load",
            cell: (row) =>
              isOverAllocated(row.overAllocated, row.warnings) ? (
                <Badge variant="destructive">Over</Badge>
              ) : (
                <Badge variant="secondary">OK</Badge>
              ),
          },
          {
            key: "actions",
            header: "",
            cell: (row) => (
              <div className="flex justify-end gap-1">
                <Button
                  size="sm"
                  variant="ghost"
                  onClick={() => {
                    setEditing(row)
                    setOpen(true)
                  }}
                >
                  Edit
                </Button>
                <Button size="sm" variant="ghost" onClick={() => setRemoving(row)}>
                  Remove
                </Button>
              </div>
            ),
          },
        ]}
        rows={list.data?.items ?? []}
        rowKey={(row) => row.id}
        isLoading={list.isLoading}
        error={list.error instanceof ApiError ? list.error.message : list.error ? "Could not load allocations" : null}
        emptyTitle="No allocations"
        emptyDescription="Allocate people to projects as a percent of their capacity."
        emptyAction={
          <Button
            onClick={() => {
              setEditing(null)
              setOpen(true)
            }}
          >
            New allocation
          </Button>
        }
        page={page}
        pageSize={DEFAULT_PAGE_SIZE}
        totalItems={list.data?.totalItems ?? 0}
        onPageChange={setPage}
      />
      <AllocationFormDialog
        open={open || Boolean(formDefaults)}
        allocation={editing}
        users={users}
        projects={projects.data?.items ?? []}
        canPickUsers={canPickUsers}
        pending={save.isPending}
        error={save.error}
        defaults={formDefaults ?? undefined}
        onOpenChange={(next) => {
          setOpen(next)
          if (!next) {
            setEditing(null)
            onFormDefaultsConsumed?.()
          }
        }}
        onSubmit={(values) => save.mutate(values)}
      />
      <AlertDialog open={Boolean(removing)} onOpenChange={(next) => !next && setRemoving(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Remove allocation?</AlertDialogTitle>
            <AlertDialogDescription>
              This does not remove {removing?.user.name} from the project. It only deletes the capacity
              assignment.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={() => removing && remove.mutate(removing.id)}
              disabled={remove.isPending}
            >
              Remove
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
