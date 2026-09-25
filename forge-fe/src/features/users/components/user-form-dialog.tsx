import { zodResolver } from "@hookform/resolvers/zod"
import { useEffect } from "react"
import { Controller, useForm } from "react-hook-form"

import { applyApiErrors } from "@/lib/validators"
import { refineUserPassword, userFormSchema, type UserFormValues } from "@/features/users/schema"
import { ApiError } from "@/types/api"
import type { Department, RoleSummary } from "@/types/common"
import type { PublicUser } from "@/types/user"
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
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Spinner } from "@/components/ui/spinner"
import { Switch } from "@/components/ui/switch"

const NONE = "none"

export function UserFormDialog({
  open,
  user,
  roles,
  departments,
  pending,
  error,
  onOpenChange,
  onSubmit,
}: {
  open: boolean
  user: PublicUser | null
  roles: RoleSummary[]
  departments: Department[]
  pending: boolean
  error: unknown
  onOpenChange: (open: boolean) => void
  onSubmit: (values: UserFormValues) => void
}) {
  const isEdit = Boolean(user)
  const form = useForm<UserFormValues>({
    resolver: zodResolver(userFormSchema.superRefine((values, ctx) => refineUserPassword(values, isEdit, ctx))),
    defaultValues: {
      name: "",
      email: "",
      password: "",
      roleId: "",
      departmentId: NONE,
      capacityHoursPerDay: 8,
      isActive: true,
      skills: "",
    },
  })

  useEffect(() => {
    if (!open) return
    form.reset({
      name: user?.name ?? "",
      email: user?.email ?? "",
      password: "",
      roleId: user?.roleId ?? "",
      departmentId: user?.departmentId ?? NONE,
      capacityHoursPerDay: user?.capacityHoursPerDay ?? 8,
      isActive: user?.isActive ?? true,
      skills: user?.skills.join(", ") ?? "",
    })
  }, [open, user, form])

  useEffect(() => {
    applyApiErrors(error, form.setError)
  }, [error, form])

  const message =
    error instanceof ApiError && error.code === "CONFLICT"
      ? "That email is already in use."
      : error instanceof ApiError && error.code === "LAST_SUPER_ADMIN"
        ? "The last Super Admin cannot be changed this way."
        : error instanceof ApiError
          ? error.message
          : null

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>{isEdit ? "Edit user" : "New user"}</DialogTitle>
          <DialogDescription>Accounts are created by an administrator. There is no public registration.</DialogDescription>
        </DialogHeader>
        <form className="flex flex-col gap-4" onSubmit={form.handleSubmit(onSubmit)}>
          <FieldGroup>
            <Field data-invalid={Boolean(form.formState.errors.name) || undefined}>
              <FieldLabel htmlFor="user-name">Name</FieldLabel>
              <Input id="user-name" aria-invalid={Boolean(form.formState.errors.name)} {...form.register("name")} />
              <FieldError errors={[form.formState.errors.name]} />
            </Field>
            <Field data-invalid={Boolean(form.formState.errors.email) || undefined}>
              <FieldLabel htmlFor="user-email">Email</FieldLabel>
              <Input
                id="user-email"
                type="email"
                disabled={isEdit}
                aria-invalid={Boolean(form.formState.errors.email)}
                {...form.register("email")}
              />
              <FieldError errors={[form.formState.errors.email]} />
            </Field>
            <Field data-invalid={Boolean(form.formState.errors.password) || undefined}>
              <FieldLabel htmlFor="user-password">{isEdit ? "New password" : "Password"}</FieldLabel>
              <Input
                id="user-password"
                type="password"
                autoComplete="new-password"
                aria-invalid={Boolean(form.formState.errors.password)}
                {...form.register("password")}
              />
              <FieldError errors={[form.formState.errors.password]} />
            </Field>
            <Field data-invalid={Boolean(form.formState.errors.roleId) || undefined}>
              <FieldLabel>Role</FieldLabel>
              <Controller
                control={form.control}
                name="roleId"
                render={({ field }) => (
                  <Select
                    items={Object.fromEntries(roles.map((role) => [role.id, role.name]))}
                    value={field.value}
                    onValueChange={(value) => field.onChange(value ?? "")}
                  >
                    <SelectTrigger className="w-full" aria-invalid={Boolean(form.formState.errors.roleId)}>
                      <SelectValue placeholder="Select a role" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectGroup>
                        {roles.map((role) => (
                          <SelectItem key={role.id} value={role.id}>
                            {role.name}
                          </SelectItem>
                        ))}
                      </SelectGroup>
                    </SelectContent>
                  </Select>
                )}
              />
              <FieldError errors={[form.formState.errors.roleId]} />
            </Field>
            <Field>
              <FieldLabel>Department</FieldLabel>
              <Controller
                control={form.control}
                name="departmentId"
                render={({ field }) => (
                  <Select
                    items={{
                      [NONE]: "No department",
                      ...Object.fromEntries(departments.map((department) => [department.id, department.name])),
                    }}
                    value={field.value}
                    onValueChange={(value) => field.onChange(value ?? NONE)}
                  >
                    <SelectTrigger className="w-full">
                      <SelectValue placeholder="No department" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectGroup>
                        <SelectItem value={NONE}>No department</SelectItem>
                        {departments.map((department) => (
                          <SelectItem key={department.id} value={department.id}>
                            {department.name}
                          </SelectItem>
                        ))}
                      </SelectGroup>
                    </SelectContent>
                  </Select>
                )}
              />
            </Field>
            <Field data-invalid={Boolean(form.formState.errors.capacityHoursPerDay) || undefined}>
              <FieldLabel htmlFor="user-capacity">Capacity hours per day</FieldLabel>
              <Input
                id="user-capacity"
                type="number"
                min={1}
                max={24}
                aria-invalid={Boolean(form.formState.errors.capacityHoursPerDay)}
                {...form.register("capacityHoursPerDay", { valueAsNumber: true })}
              />
              <FieldError errors={[form.formState.errors.capacityHoursPerDay]} />
            </Field>
            <Field>
              <FieldLabel htmlFor="user-skills">Skills</FieldLabel>
              <Input id="user-skills" placeholder="network, golang" {...form.register("skills")} />
            </Field>
            <Field orientation="horizontal">
              <Controller
                control={form.control}
                name="isActive"
                render={({ field }) => (
                  <Switch id="user-active" checked={field.value} onCheckedChange={field.onChange} />
                )}
              />
              <FieldLabel htmlFor="user-active">Active</FieldLabel>
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
