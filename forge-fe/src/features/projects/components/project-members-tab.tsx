import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { useState } from "react"
import { toast } from "sonner"

import { EntityTable } from "@/components/common/entity-table"
import { Button } from "@/components/ui/button"
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
import { addProjectMember, listProjectMembers, removeProjectMember } from "@/features/projects/api/members"
import { AddMemberDialog, type AddMemberValues } from "@/features/projects/components/add-member-dialog"
import { listUsers } from "@/features/users/api/users"
import { can, formatRole } from "@/lib/auth"
import { DEFAULT_PAGE_SIZE, PERMISSIONS } from "@/lib/constants"
import { useSession } from "@/lib/session"
import { ApiError } from "@/types/api"
import type { ProjectMember, ProjectPublic } from "@/types/project"
import { queryKeys } from "@/services/query/query-keys"

export function ProjectMembersTab({
  project,
  canManage,
}: {
  project: ProjectPublic
  canManage: boolean
}) {
  const queryClient = useQueryClient()
  const permissions = useSession((state) => state.permissions)
  const canPickUsers = can(permissions, PERMISSIONS.userManage)
  const [page, setPage] = useState(1)
  const [open, setOpen] = useState(false)
  const [removing, setRemoving] = useState<ProjectMember | null>(null)

  const list = useQuery({
    queryKey: queryKeys.projectMembers(project.id, { page, pageSize: DEFAULT_PAGE_SIZE }),
    queryFn: () => listProjectMembers(project.id, page, DEFAULT_PAGE_SIZE),
  })

  const users = useQuery({
    queryKey: queryKeys.users({ page: 1, pageSize: 100, picker: true }),
    queryFn: () => listUsers({ page: 1, pageSize: 100 }),
    enabled: canPickUsers && open,
  })

  const add = useMutation({
    mutationFn: (values: AddMemberValues) => addProjectMember(project.id, values.userId, values.role),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["projects", project.id] })
      toast.success("Member added")
      setOpen(false)
    },
    onError: (error) => {
      toast.error(error instanceof ApiError ? error.message : "Could not add member")
    },
  })

  const remove = useMutation({
    mutationFn: (userId: string) => removeProjectMember(project.id, userId),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["projects", project.id] })
      toast.success("Member removed")
      setRemoving(null)
    },
    onError: (error) => {
      toast.error(error instanceof ApiError ? error.message : "Could not remove member")
    },
  })

  const error =
    list.error instanceof ApiError ? list.error.message : list.error ? "Could not load members" : null

  return (
    <div className="flex flex-col gap-3">
      {canManage ? (
        <div className="flex justify-end">
          <Button
            size="sm"
            onClick={() => {
              add.reset()
              setOpen(true)
            }}
          >
            Add member
          </Button>
        </div>
      ) : null}
      <EntityTable
        columns={[
          { key: "name", header: "Name", cell: (row) => row.name },
          { key: "department", header: "Department", cell: (row) => row.departmentName ?? "—" },
          { key: "role", header: "Project role", cell: (row) => formatRole(row.projectRole) },
          {
            key: "actions",
            header: "",
            cell: (row) =>
              canManage ? (
                <div className="flex justify-end">
                  <Button variant="outline" size="sm" onClick={() => setRemoving(row)}>
                    Remove
                  </Button>
                </div>
              ) : null,
          },
        ]}
        rows={list.data?.items ?? []}
        rowKey={(row) => row.userId}
        isLoading={list.isLoading}
        error={error}
        emptyTitle="No members"
        emptyDescription="Add people with a project role."
        page={page}
        pageSize={list.data?.pageSize ?? DEFAULT_PAGE_SIZE}
        totalItems={list.data?.totalItems ?? 0}
        onPageChange={setPage}
      />
      <AddMemberDialog
        open={open}
        users={users.data?.items ?? []}
        canPickUsers={canPickUsers}
        pending={add.isPending}
        error={add.error}
        onOpenChange={setOpen}
        onSubmit={(values) => add.mutate(values)}
      />
      <AlertDialog open={Boolean(removing)} onOpenChange={(next) => !next && setRemoving(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Remove member</AlertDialogTitle>
            <AlertDialogDescription>
              {removing ? `Remove ${removing.name} from this project?` : ""}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              variant="destructive"
              disabled={remove.isPending}
              onClick={() => removing && remove.mutate(removing.userId)}
            >
              Remove
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
