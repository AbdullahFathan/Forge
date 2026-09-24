import { zodResolver } from "@hookform/resolvers/zod"
import { useEffect } from "react"
import { Controller, useForm } from "react-hook-form"

import { applyApiErrors } from "@/lib/validators"
import { roleFormSchema, type RoleFormValues } from "@/features/roles/schema"
import { ApiError } from "@/types/api"
import type { Permission, RoleSummary } from "@/types/common"
import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import {
  Field,
  FieldError,
  FieldGroup,
  FieldLabel,
  FieldLegend,
  FieldSet,
} from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { Spinner } from "@/components/ui/spinner"

export function RoleFormDialog({
  open,
  role,
  permissions,
  permissionsLoading,
  permissionsError,
  pending,
  error,
  onOpenChange,
  onSubmit,
}: {
  open: boolean
  role: RoleSummary | null
  permissions: Permission[]
  permissionsLoading: boolean
  permissionsError: unknown
  pending: boolean
  error: unknown
  onOpenChange: (open: boolean) => void
  onSubmit: (values: RoleFormValues) => void
}) {
  const form = useForm<RoleFormValues>({
    resolver: zodResolver(roleFormSchema),
    defaultValues: { name: "", permissionCodes: [] },
  })

  useEffect(() => {
    if (!open) return
    form.reset({
      name: role?.name ?? "",
      permissionCodes: role?.permissionCodes ?? [],
    })
  }, [open, role, form])

  useEffect(() => {
    applyApiErrors(error, form.setError)
  }, [error, form])

  const apiMessage =
    error instanceof ApiError
      ? error.code === "FORBIDDEN"
        ? "System roles cannot be changed."
        : error.code === "CONFLICT"
          ? error.message || "This role code is already in use."
          : error.message
      : null

  const loadError =
    permissionsError instanceof ApiError
      ? permissionsError.message
      : permissionsError
        ? "Could not load permissions."
        : null

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>{role ? "Edit role" : "New role"}</DialogTitle>
          <DialogDescription>
            Custom roles are a named set of permissions. System roles cannot be edited here.
          </DialogDescription>
        </DialogHeader>
        <form className="flex flex-col gap-4" onSubmit={form.handleSubmit(onSubmit)}>
          <FieldGroup>
            <Field data-invalid={Boolean(form.formState.errors.name) || undefined}>
              <FieldLabel htmlFor="role-name">Name</FieldLabel>
              <Input
                id="role-name"
                aria-invalid={Boolean(form.formState.errors.name)}
                {...form.register("name")}
              />
              <FieldError errors={[form.formState.errors.name]} />
            </Field>
            <FieldSet>
              <FieldLegend>Permissions</FieldLegend>
              <Field data-invalid={Boolean(form.formState.errors.permissionCodes) || undefined}>
                {permissionsLoading ? (
                  <p className="text-sm text-muted-foreground">Loading permissions…</p>
                ) : loadError ? (
                  <p className="text-sm text-destructive">{loadError}</p>
                ) : (
                  <Controller
                    control={form.control}
                    name="permissionCodes"
                    render={({ field }) => (
                      <div className="flex max-h-60 flex-col gap-2 overflow-auto rounded-lg border p-2">
                        {permissions.map((permission) => {
                          const checked = field.value.includes(permission.code)
                          return (
                            <Field key={permission.id} orientation="horizontal">
                              <Checkbox
                                checked={checked}
                                onCheckedChange={(value) => {
                                  const next = value
                                    ? [...field.value, permission.code]
                                    : field.value.filter((code) => code !== permission.code)
                                  field.onChange(next)
                                }}
                              />
                              <FieldLabel>
                                {permission.name}
                                <span className="ml-1 text-muted-foreground">({permission.code})</span>
                              </FieldLabel>
                            </Field>
                          )
                        })}
                      </div>
                    )}
                  />
                )}
                <FieldError errors={[form.formState.errors.permissionCodes]} />
              </Field>
            </FieldSet>
          </FieldGroup>
          {apiMessage ? <p className="text-sm text-destructive">{apiMessage}</p> : null}
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={pending || permissionsLoading}>
              {pending ? <Spinner data-icon="inline-start" /> : null}
              Save
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
