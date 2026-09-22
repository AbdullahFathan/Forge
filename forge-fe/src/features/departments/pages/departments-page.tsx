import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { useState } from "react"
import { toast } from "sonner"

import { EntityTable } from "@/components/common/entity-table"
import { PageHeader } from "@/components/common/page-header"
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
import { DepartmentFormDialog } from "@/features/departments/components/department-form-dialog"
import {
  createDepartment,
  deleteDepartment,
  listDepartments,
  patchDepartment,
} from "@/features/departments/api/departments"
import type { DepartmentValues } from "@/features/departments/schema"
import { DEFAULT_PAGE_SIZE } from "@/lib/constants"
import { ApiError } from "@/types/api"
import type { Department } from "@/types/common"
import { queryKeys } from "@/services/query/query-keys"

export function DepartmentsPage() {
  const queryClient = useQueryClient()
  const [page, setPage] = useState(1)
  const [open, setOpen] = useState(false)
  const [editing, setEditing] = useState<Department | null>(null)
  const [removing, setRemoving] = useState<Department | null>(null)

  const list = useQuery({
    queryKey: queryKeys.departments({ page, pageSize: DEFAULT_PAGE_SIZE }),
    queryFn: () => listDepartments(page, DEFAULT_PAGE_SIZE),
  })

  const save = useMutation({
    mutationFn: (values: DepartmentValues) =>
      editing ? patchDepartment(editing.id, values.name) : createDepartment(values.name),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["departments"] })
      toast.success(editing ? "Department updated" : "Department created")
      setOpen(false)
      setEditing(null)
    },
  })

  const remove = useMutation({
    mutationFn: (id: string) => deleteDepartment(id),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["departments"] })
      toast.success("Department removed")
      setRemoving(null)
    },
  })

  const error =
    list.error instanceof ApiError
      ? list.error.message
      : list.error
        ? "Could not load departments"
        : null

  return (
    <div className="flex flex-col gap-4">
      <PageHeader
        title="Departments"
        description="Organization units used on users and projects."
        actions={
          <Button
            onClick={() => {
              setEditing(null)
              save.reset()
              setOpen(true)
            }}
          >
            New department
          </Button>
        }
      />
      <EntityTable
        columns={[
          { key: "name", header: "Name", cell: (row) => row.name },
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
        emptyTitle="No departments yet"
        emptyDescription="Create a department before assigning people."
        emptyAction={
          <Button
            onClick={() => {
              setEditing(null)
              setOpen(true)
            }}
          >
            New department
          </Button>
        }
        page={page}
        pageSize={list.data?.pageSize ?? DEFAULT_PAGE_SIZE}
        totalItems={list.data?.totalItems ?? 0}
        onPageChange={setPage}
      />
      <DepartmentFormDialog
        open={open}
        department={editing}
        pending={save.isPending}
        error={save.error}
        onOpenChange={setOpen}
        onSubmit={(values) => save.mutate(values)}
      />
      <AlertDialog open={Boolean(removing)} onOpenChange={(next) => !next && setRemoving(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Remove department</AlertDialogTitle>
            <AlertDialogDescription>
              {removing ? `Remove ${removing.name}? People keep their accounts.` : ""}
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
