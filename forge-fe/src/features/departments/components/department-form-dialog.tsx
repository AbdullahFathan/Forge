import { zodResolver } from "@hookform/resolvers/zod"
import { useEffect } from "react"
import { useForm } from "react-hook-form"

import { applyApiErrors } from "@/lib/validators"
import { departmentSchema, type DepartmentValues } from "@/features/departments/schema"
import { ApiError } from "@/types/api"
import type { Department } from "@/types/common"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Field, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { Spinner } from "@/components/ui/spinner"

export function DepartmentFormDialog({
  open,
  department,
  pending,
  error,
  onOpenChange,
  onSubmit,
}: {
  open: boolean
  department: Department | null
  pending: boolean
  error: unknown
  onOpenChange: (open: boolean) => void
  onSubmit: (values: DepartmentValues) => void
}) {
  const form = useForm<DepartmentValues>({
    resolver: zodResolver(departmentSchema),
    defaultValues: { name: "" },
  })

  useEffect(() => {
    if (!open) return
    form.reset({ name: department?.name ?? "" })
  }, [open, department, form])

  useEffect(() => {
    applyApiErrors(error, form.setError)
  }, [error, form])

  const message = error instanceof ApiError ? error.message : null

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>{department ? "Edit department" : "New department"}</DialogTitle>
          <DialogDescription>Departments group people for capacity and reporting.</DialogDescription>
        </DialogHeader>
        <form className="flex flex-col gap-4" onSubmit={form.handleSubmit(onSubmit)}>
          <FieldGroup>
            <Field data-invalid={Boolean(form.formState.errors.name) || undefined}>
              <FieldLabel htmlFor="department-name">Name</FieldLabel>
              <Input
                id="department-name"
                aria-invalid={Boolean(form.formState.errors.name)}
                {...form.register("name")}
              />
              <FieldError errors={[form.formState.errors.name]} />
            </Field>
          </FieldGroup>
          {message ? <p className="text-sm text-destructive">{message}</p> : null}
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={pending}>
              {pending ? <Spinner data-icon="inline-start" /> : null}
              Save
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
