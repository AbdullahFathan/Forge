import { zodResolver } from "@hookform/resolvers/zod"
import { useEffect } from "react"
import { Controller, useForm } from "react-hook-form"

import { applyApiErrors } from "@/lib/validators"
import { toIsoDate } from "@/lib/date-range"
import {
  ALLOCATION_ROLES,
  allocationFormSchema,
  type AllocationFormValues,
} from "@/features/resources/schema"
import { ApiError } from "@/types/api"
import type { AllocationPublic } from "@/types/resource"
import type { ProjectPublic } from "@/types/project"
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

export function AllocationFormDialog({
  open,
  allocation,
  users,
  projects,
  canPickUsers,
  pending,
  error,
  defaults,
  onOpenChange,
  onSubmit,
}: {
  open: boolean
  allocation: AllocationPublic | null
  users: PublicUser[]
  projects: ProjectPublic[]
  canPickUsers: boolean
  pending: boolean
  error: unknown
  defaults?: Partial<AllocationFormValues>
  onOpenChange: (open: boolean) => void
  onSubmit: (values: AllocationFormValues) => void
}) {
  const isEdit = Boolean(allocation)
  const form = useForm<AllocationFormValues>({
    resolver: zodResolver(allocationFormSchema),
    defaultValues: {
      userId: "",
      projectId: "",
      allocationPercent: 50,
      startDate: toIsoDate(new Date()),
      endDate: toIsoDate(new Date()),
      role: "MEMBER",
    },
  })

  useEffect(() => {
    if (!open) return
    form.reset({
      userId: allocation?.userId ?? defaults?.userId ?? "",
      projectId: allocation?.projectId ?? defaults?.projectId ?? "",
      allocationPercent: allocation?.allocationPercent ?? defaults?.allocationPercent ?? 50,
      startDate: allocation?.startDate ?? defaults?.startDate ?? toIsoDate(new Date()),
      endDate: allocation?.endDate ?? defaults?.endDate ?? toIsoDate(new Date()),
      role: (allocation?.role as AllocationFormValues["role"]) ?? defaults?.role ?? "MEMBER",
    })
  }, [open, allocation, defaults, form])

  useEffect(() => {
    applyApiErrors(error, form.setError)
  }, [error, form])

  const message =
    error instanceof ApiError && error.code === "CONFLICT"
      ? "This person already has an overlapping allocation on the same project."
      : error instanceof ApiError
        ? error.message
        : null

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>{isEdit ? "Edit allocation" : "New allocation"}</DialogTitle>
          <DialogDescription>
            Saving adds the person as a project member (or updates their project role). Removing an
            allocation does not remove them from the project.
          </DialogDescription>
        </DialogHeader>
        <form className="flex flex-col gap-4" onSubmit={form.handleSubmit(onSubmit)}>
          <FieldGroup>
            <Field data-invalid={Boolean(form.formState.errors.userId) || undefined}>
              <FieldLabel htmlFor="alloc-user">Person</FieldLabel>
              {canPickUsers ? (
                <Controller
                  control={form.control}
                  name="userId"
                  render={({ field }) => (
                    <Select
                      value={field.value}
                      onValueChange={(value) => field.onChange(value ?? "")}
                      disabled={isEdit}
                    >
                      <SelectTrigger className="w-full" aria-invalid={Boolean(form.formState.errors.userId)}>
                        <SelectValue placeholder="Select a person" />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectGroup>
                          {users.map((user) => (
                            <SelectItem key={user.id} value={user.id}>
                              {user.name}
                            </SelectItem>
                          ))}
                        </SelectGroup>
                      </SelectContent>
                    </Select>
                  )}
                />
              ) : (
                <Input
                  id="alloc-user"
                  disabled={isEdit}
                  placeholder="User ID"
                  aria-invalid={Boolean(form.formState.errors.userId)}
                  {...form.register("userId")}
                />
              )}
              {!canPickUsers && !isEdit ? (
                <p className="text-xs text-muted-foreground">
                  User list requires user.manage. Paste a user ID, or ask an administrator.
                </p>
              ) : null}
              <FieldError errors={[form.formState.errors.userId]} />
            </Field>
            <Field data-invalid={Boolean(form.formState.errors.projectId) || undefined}>
              <FieldLabel>Project</FieldLabel>
              <Controller
                control={form.control}
                name="projectId"
                render={({ field }) => (
                  <Select
                    value={field.value}
                    onValueChange={(value) => field.onChange(value ?? "")}
                    disabled={isEdit}
                  >
                    <SelectTrigger className="w-full" aria-invalid={Boolean(form.formState.errors.projectId)}>
                      <SelectValue placeholder="Select a project" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectGroup>
                        {projects.map((project) => (
                          <SelectItem key={project.id} value={project.id}>
                            {project.name}
                          </SelectItem>
                        ))}
                      </SelectGroup>
                    </SelectContent>
                  </Select>
                )}
              />
              <FieldError errors={[form.formState.errors.projectId]} />
            </Field>
            <Field data-invalid={Boolean(form.formState.errors.allocationPercent) || undefined}>
              <FieldLabel htmlFor="alloc-percent">Allocation %</FieldLabel>
              <Input
                id="alloc-percent"
                type="number"
                min={0}
                max={100}
                aria-invalid={Boolean(form.formState.errors.allocationPercent)}
                {...form.register("allocationPercent", { valueAsNumber: true })}
              />
              <FieldError errors={[form.formState.errors.allocationPercent]} />
            </Field>
            <Field data-invalid={Boolean(form.formState.errors.startDate) || undefined}>
              <FieldLabel htmlFor="alloc-start">Start</FieldLabel>
              <Input
                id="alloc-start"
                type="date"
                aria-invalid={Boolean(form.formState.errors.startDate)}
                {...form.register("startDate")}
              />
              <FieldError errors={[form.formState.errors.startDate]} />
            </Field>
            <Field data-invalid={Boolean(form.formState.errors.endDate) || undefined}>
              <FieldLabel htmlFor="alloc-end">End</FieldLabel>
              <Input
                id="alloc-end"
                type="date"
                aria-invalid={Boolean(form.formState.errors.endDate)}
                {...form.register("endDate")}
              />
              <FieldError errors={[form.formState.errors.endDate]} />
            </Field>
            <Field data-invalid={Boolean(form.formState.errors.role) || undefined}>
              <FieldLabel>Project role</FieldLabel>
              <Controller
                control={form.control}
                name="role"
                render={({ field }) => (
                  <Select value={field.value} onValueChange={(value) => field.onChange(value ?? "MEMBER")}>
                    <SelectTrigger className="w-full">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectGroup>
                        {ALLOCATION_ROLES.map((role) => (
                          <SelectItem key={role} value={role}>
                            {role}
                          </SelectItem>
                        ))}
                      </SelectGroup>
                    </SelectContent>
                  </Select>
                )}
              />
              <FieldError errors={[form.formState.errors.role]} />
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
