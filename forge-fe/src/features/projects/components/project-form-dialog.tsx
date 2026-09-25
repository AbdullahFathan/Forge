import { zodResolver } from "@hookform/resolvers/zod"
import { useEffect } from "react"
import { Controller, useForm } from "react-hook-form"

import { applyApiErrors } from "@/lib/validators"
import { nextProjectStatuses } from "@/lib/lifecycle"
import { projectStatusMap, priorityMap } from "@/components/common/status-map"
import {
  createStatusOptions,
  projectFormSchema,
  type ProjectFormValues,
} from "@/features/projects/schema"
import { ApiError } from "@/types/api"
import type { Department } from "@/types/common"
import type { ProjectPublic } from "@/types/project"
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
import { Textarea } from "@/components/ui/textarea"
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Spinner } from "@/components/ui/spinner"

const NONE = "none"

export function ProjectFormDialog({
  open,
  project,
  departments,
  canPickDepartment,
  pending,
  error,
  onOpenChange,
  onSubmit,
}: {
  open: boolean
  project: ProjectPublic | null
  departments: Department[]
  canPickDepartment: boolean
  pending: boolean
  error: unknown
  onOpenChange: (open: boolean) => void
  onSubmit: (values: ProjectFormValues) => void
}) {
  const isEdit = Boolean(project)
  const statusOptions = isEdit && project ? nextProjectStatuses(project.status) : [...createStatusOptions]
  const form = useForm<ProjectFormValues>({
    resolver: zodResolver(projectFormSchema),
    defaultValues: {
      name: "",
      description: "",
      status: "DRAFT",
      priority: "MEDIUM",
      startDate: "",
      targetEndDate: "",
      departmentId: NONE,
      tags: "",
    },
  })

  useEffect(() => {
    if (!open) return
    form.reset({
      name: project?.name ?? "",
      description: project?.description ?? "",
      status: (project?.status as ProjectFormValues["status"]) ?? "DRAFT",
      priority: (project?.priority as ProjectFormValues["priority"]) ?? "MEDIUM",
      startDate: project?.startDate ?? "",
      targetEndDate: project?.targetEndDate ?? "",
      departmentId: project?.departmentId ?? NONE,
      tags: project?.tags.join(", ") ?? "",
    })
  }, [open, project, form])

  useEffect(() => {
    applyApiErrors(error, form.setError)
  }, [error, form])

  const message = error instanceof ApiError ? error.message : null

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>{isEdit ? "Edit project" : "New project"}</DialogTitle>
          <DialogDescription>
            {isEdit ? "Update fields the API allows. Illegal status jumps are rejected." : "Creates a project you own. Archive is used later instead of delete."}
          </DialogDescription>
        </DialogHeader>
        <form className="flex flex-col gap-4" onSubmit={form.handleSubmit(onSubmit)}>
          <FieldGroup>
            <Field data-invalid={Boolean(form.formState.errors.name) || undefined}>
              <FieldLabel htmlFor="project-name">Name</FieldLabel>
              <Input id="project-name" aria-invalid={Boolean(form.formState.errors.name)} {...form.register("name")} />
              <FieldError errors={[form.formState.errors.name]} />
            </Field>
            <Field>
              <FieldLabel htmlFor="project-description">Description</FieldLabel>
              <Textarea id="project-description" rows={3} {...form.register("description")} />
            </Field>
            <Field data-invalid={Boolean(form.formState.errors.status) || undefined}>
              <FieldLabel>Status</FieldLabel>
              <Controller
                control={form.control}
                name="status"
                render={({ field }) => (
                  <Select value={field.value} onValueChange={(value) => field.onChange(value ?? "DRAFT")}>
                    <SelectTrigger className="w-full" aria-invalid={Boolean(form.formState.errors.status)}>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectGroup>
                        {statusOptions.map((status) => (
                          <SelectItem key={status} value={status}>
                            {projectStatusMap[status]?.label ?? status}
                          </SelectItem>
                        ))}
                      </SelectGroup>
                    </SelectContent>
                  </Select>
                )}
              />
              <FieldError errors={[form.formState.errors.status]} />
            </Field>
            <Field>
              <FieldLabel>Priority</FieldLabel>
              <Controller
                control={form.control}
                name="priority"
                render={({ field }) => (
                  <Select value={field.value} onValueChange={(value) => field.onChange(value ?? "MEDIUM")}>
                    <SelectTrigger className="w-full">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectGroup>
                        {Object.entries(priorityMap).map(([value, visual]) => (
                          <SelectItem key={value} value={value}>
                            {visual.label}
                          </SelectItem>
                        ))}
                      </SelectGroup>
                    </SelectContent>
                  </Select>
                )}
              />
            </Field>
            <Field data-invalid={Boolean(form.formState.errors.startDate) || undefined}>
              <FieldLabel htmlFor="project-start">Start date</FieldLabel>
              <Input
                id="project-start"
                type="date"
                aria-invalid={Boolean(form.formState.errors.startDate)}
                {...form.register("startDate")}
              />
              <FieldError errors={[form.formState.errors.startDate]} />
            </Field>
            <Field data-invalid={Boolean(form.formState.errors.targetEndDate) || undefined}>
              <FieldLabel htmlFor="project-end">Target end date</FieldLabel>
              <Input
                id="project-end"
                type="date"
                aria-invalid={Boolean(form.formState.errors.targetEndDate)}
                {...form.register("targetEndDate")}
              />
              <FieldError errors={[form.formState.errors.targetEndDate]} />
            </Field>
            {canPickDepartment ? (
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
            ) : null}
            <Field>
              <FieldLabel htmlFor="project-tags">Tags</FieldLabel>
              <Input id="project-tags" placeholder="infra, q3" {...form.register("tags")} />
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
