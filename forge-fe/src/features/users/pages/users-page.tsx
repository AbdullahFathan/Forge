import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { useState } from "react"
import { toast } from "sonner"

import { EntityTable } from "@/components/common/entity-table"
import { FilterBar } from "@/components/common/filter-bar"
import { PageHeader } from "@/components/common/page-header"
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
import { Badge } from "@/components/ui/badge"
import { listDepartments } from "@/features/departments/api/departments"
import { listRoles } from "@/features/roles/api/roles"
import { createUser, deleteUser, listUsers, patchUser } from "@/features/users/api/users"
import { UserFormDialog } from "@/features/users/components/user-form-dialog"
import { toUserPayload, type UserFormValues } from "@/features/users/schema"
import { formatRole } from "@/lib/auth"
import { DEFAULT_PAGE_SIZE } from "@/lib/constants"
import { ApiError } from "@/types/api"
import type { PublicUser } from "@/types/user"
import { queryKeys } from "@/services/query/query-keys"

const ALL = "all"

export function UsersPage() {
  const queryClient = useQueryClient()
  const [page, setPage] = useState(1)
  const [roleId, setRoleId] = useState(ALL)
  const [departmentId, setDepartmentId] = useState(ALL)
  const [open, setOpen] = useState(false)
  const [editing, setEditing] = useState<PublicUser | null>(null)
  const [removing, setRemoving] = useState<PublicUser | null>(null)

  const roles = useQuery({ queryKey: queryKeys.roles, queryFn: listRoles })
  const departments = useQuery({
    queryKey: queryKeys.departments({ page: 1, pageSize: 100, picker: true }),
    queryFn: () => listDepartments(1, 100),
  })

  const params = {
    page,
    pageSize: DEFAULT_PAGE_SIZE,
    ...(roleId !== ALL ? { roleId } : {}),
    ...(departmentId !== ALL ? { departmentId } : {}),
  }

  const list = useQuery({
    queryKey: queryKeys.users(params),
    queryFn: () => listUsers(params),
  })

  const save = useMutation({
    mutationFn: (values: UserFormValues) => {
      const payload = toUserPayload(values, Boolean(editing))
      return editing ? patchUser(editing.id, payload) : createUser(payload)
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["users"] })
      toast.success(editing ? "User updated" : "User created")
      setOpen(false)
      setEditing(null)
    },
  })

  const remove = useMutation({
    mutationFn: (id: string) => deleteUser(id),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["users"] })
      toast.success("User removed")
      setRemoving(null)
    },
    onError: (error) => {
      toast.error(error instanceof ApiError ? error.message : "Could not remove user")
    },
  })

  const error =
    list.error instanceof ApiError ? list.error.message : list.error ? "Could not load users" : null

  return (
    <div className="flex flex-col gap-4">
      <PageHeader
        title="Users"
        description="People who can sign in. Passwords are never shown after they are set."
        actions={
          <Button
            onClick={() => {
              setEditing(null)
              save.reset()
              setOpen(true)
            }}
          >
            New user
          </Button>
        }
      />
      <FilterBar>
        <Select
          items={{
            [ALL]: "All roles",
            ...Object.fromEntries((roles.data ?? []).map((role) => [role.id, role.name])),
          }}
          value={roleId}
          onValueChange={(value) => {
            setRoleId(value ?? ALL)
            setPage(1)
          }}
        >
          <SelectTrigger className="w-44" aria-label="Filter by role">
            <SelectValue placeholder="All roles" />
          </SelectTrigger>
          <SelectContent>
            <SelectGroup>
              <SelectItem value={ALL}>All roles</SelectItem>
              {(roles.data ?? []).map((role) => (
                <SelectItem key={role.id} value={role.id}>
                  {role.name}
                </SelectItem>
              ))}
            </SelectGroup>
          </SelectContent>
        </Select>
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
      </FilterBar>
      <EntityTable
        columns={[
          { key: "name", header: "Name", cell: (row) => row.name },
          { key: "email", header: "Email", cell: (row) => row.email },
          { key: "role", header: "Role", cell: (row) => formatRole(row.roleCode) },
          { key: "department", header: "Department", cell: (row) => row.departmentName ?? "—" },
          {
            key: "capacity",
            header: "Hours / day",
            className: "px-3 py-2 font-mono tabular-nums",
            cell: (row) => row.capacityHoursPerDay,
          },
          {
            key: "active",
            header: "Status",
            cell: (row) => (
              <Badge variant={row.isActive ? "default" : "secondary"}>
                {row.isActive ? "Active" : "Inactive"}
              </Badge>
            ),
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
                <Button variant="outline" size="sm" onClick={() => setRemoving(row)}>
                  Remove
                </Button>
              </div>
            ),
          },
        ]}
        rows={list.data?.items ?? []}
        rowKey={(row) => row.id}
        isLoading={list.isLoading}
        error={error}
        emptyTitle="No users match"
        emptyDescription="Adjust the filters or create a user."
        page={page}
        pageSize={list.data?.pageSize ?? DEFAULT_PAGE_SIZE}
        totalItems={list.data?.totalItems ?? 0}
        onPageChange={setPage}
      />
      <UserFormDialog
        open={open}
        user={editing}
        roles={roles.data ?? []}
        departments={departments.data?.items ?? []}
        pending={save.isPending}
        error={save.error}
        onOpenChange={setOpen}
        onSubmit={(values) => save.mutate(values)}
      />
      <AlertDialog open={Boolean(removing)} onOpenChange={(next) => !next && setRemoving(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Remove user</AlertDialogTitle>
            <AlertDialogDescription>
              {removing ? `Remove ${removing.name}? This is a soft delete.` : ""}
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
