import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { useState } from "react"
import { toast } from "sonner"

import { EntityTable } from "@/components/common/entity-table"
import { PageHeader } from "@/components/common/page-header"
import { Badge } from "@/components/ui/badge"
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
import {
  createRole,
  deleteRole,
  listPermissions,
  listRoles,
  patchRole,
} from "@/features/roles/api/roles"
import { RoleFormDialog } from "@/features/roles/components/role-form-dialog"
import { roleActionsVisible } from "@/features/roles/role-actions"
import type { RoleFormValues } from "@/features/roles/schema"
import { ApiError } from "@/types/api"
import type { RoleSummary } from "@/types/common"
import { queryKeys } from "@/services/query/query-keys"

function mutationMessage(error: unknown, fallback: string) {
  if (!(error instanceof ApiError)) return fallback
  if (error.code === "FORBIDDEN") return "System roles cannot be changed."
  if (error.code === "CONFLICT") {
    return error.message.includes("assigned")
      ? "This role is still assigned to users."
      : error.message || "This role code is already in use."
  }
  return error.message || fallback
}

export function RolesPage() {
  const queryClient = useQueryClient()
  const [open, setOpen] = useState(false)
  const [editing, setEditing] = useState<RoleSummary | null>(null)
  const [removing, setRemoving] = useState<RoleSummary | null>(null)

  const list = useQuery({
    queryKey: queryKeys.roles,
    queryFn: listRoles,
  })

  const permissions = useQuery({
    queryKey: queryKeys.permissions,
    queryFn: listPermissions,
  })

  const save = useMutation({
    mutationFn: (values: RoleFormValues) =>
      editing ? patchRole(editing.id, values) : createRole(values),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: queryKeys.roles })
      toast.success(editing ? "Role updated" : "Role created")
      setOpen(false)
      setEditing(null)
    },
  })

  const remove = useMutation({
    mutationFn: (id: string) => deleteRole(id),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: queryKeys.roles })
      toast.success("Role removed")
      setRemoving(null)
    },
    onError: (error) => {
      toast.error(mutationMessage(error, "Could not remove role"))
    },
  })

  const rows = list.data ?? []
  const error =
    list.error instanceof ApiError
      ? list.error.message
      : list.error
        ? "Could not load roles"
        : null

  return (
    <div className="flex flex-col gap-4">
      <PageHeader
        title="Roles"
        description="Custom roles pick a set of permissions. System roles cannot be edited or removed."
        actions={
          <Button
            onClick={() => {
              setEditing(null)
              save.reset()
              setOpen(true)
            }}
          >
            New role
          </Button>
        }
      />
      <EntityTable
        columns={[
          { key: "name", header: "Name", cell: (row) => row.name },
          { key: "code", header: "Code", cell: (row) => row.code },
          {
            key: "system",
            header: "Type",
            cell: (row) =>
              row.isSystem ? <Badge variant="secondary">System</Badge> : <Badge variant="outline">Custom</Badge>,
          },
          {
            key: "permissions",
            header: "Permissions",
            cell: (row) => `${row.permissionCodes.length} permissions`,
          },
          {
            key: "actions",
            header: "",
            cell: (row) =>
              roleActionsVisible(row) ? (
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
              ) : null,
          },
        ]}
        rows={rows}
        rowKey={(row) => row.id}
        isLoading={list.isLoading}
        error={error}
        emptyTitle="No roles yet"
        emptyDescription="Create a custom role, or wait for system roles to load."
        emptyAction={
          <Button
            onClick={() => {
              setEditing(null)
              setOpen(true)
            }}
          >
            New role
          </Button>
        }
        page={1}
        pageSize={Math.max(rows.length, 1)}
        totalItems={rows.length}
        onPageChange={() => undefined}
      />
      <RoleFormDialog
        open={open}
        role={editing}
        permissions={permissions.data ?? []}
        permissionsLoading={permissions.isLoading}
        permissionsError={permissions.error}
        pending={save.isPending}
        error={save.error}
        onOpenChange={setOpen}
        onSubmit={(values) => save.mutate(values)}
      />
      <AlertDialog open={Boolean(removing)} onOpenChange={(next) => !next && setRemoving(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Remove role?</AlertDialogTitle>
            <AlertDialogDescription>
              {removing
                ? `Remove ${removing.name}? This only works if no one is assigned the role.`
                : ""}
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
