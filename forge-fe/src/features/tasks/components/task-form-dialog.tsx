import { zodResolver } from "@hookform/resolvers/zod"
import { useEffect } from "react"
import { Controller, useForm } from "react-hook-form"

import { applyApiErrors } from "@/lib/validators"
import { nextTaskStatuses } from "@/lib/lifecycle"
import { priorityMap, taskStatusMap } from "@/components/common/status-map"
import { taskFormSchema, type TaskFormValues } from "@/features/tasks/schema"
import { ApiError } from "@/types/api"
import type { ProjectMember } from "@/types/project"
import type { TaskPublic } from "@/types/task"
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

export function TaskFormDialog({
  open,
  task,
  parentName,
  members,
  pending,
  error,
  onOpenChange,
  onSubmit,
}: {
  open: boolean
  task: TaskPublic | null
  parentName?: string | null
  members: ProjectMember[]
  pending: boolean
  error: unknown
  onOpenChange: (open: boolean) => void
  onSubmit: (values: TaskFormValues) => void
}) {
  const isEdit = Boolean(task)
  const statusOptions = isEdit && task ? nextTaskStatuses(task.status) : ["BACKLOG", "TODO"]
  const form = useForm<TaskFormValues>({
    resolver: zodResolver(taskFormSchema),
    defaultValues: {
      name: "",
      description: "",
      status: "BACKLOG",
      priority: "MEDIUM",
      estimatedHours: "",
      startDate: "",
      dueDate: "",
      labels: "",
      assigneeIds: [],
    },
  })

  useEffect(() => {
    if (!open) return
    form.reset({
      name: task?.name ?? "",
      description: task?.description ?? "",
      status: (task?.status as TaskFormValues["status"]) ?? "BACKLOG",
      priority: (task?.priority as TaskFormValues["priority"]) ?? "MEDIUM",
      estimatedHours: task?.estimatedHours != null ? String(task.estimatedHours) : "",
      startDate: task?.startDate ?? "",
      dueDate: task?.dueDate ?? "",
      labels: task?.labels.join(", ") ?? "",
      assigneeIds: task?.assignees.map((person) => person.id) ?? [],
    })
  }, [open, task, form])

  useEffect(() => {
    applyApiErrors(error, form.setError)
  }, [error, form])

  const message = error instanceof ApiError ? error.message : null
  const title = isEdit ? "Edit task" : parentName ? `Subtask of ${parentName}` : "Create New Task"

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent
        overlayClassName="bg-black/40 supports-backdrop-filter:backdrop-blur-sm"
        className="max-h-[min(90vh,44rem)] overflow-y-auto sm:max-w-4xl sm:p-6"
      >
        <DialogHeader>
          <DialogTitle className="text-lg">{title}</DialogTitle>
          <DialogDescription>Assignees must already be project members.</DialogDescription>
        </DialogHeader>
        <form className="flex flex-col gap-5" onSubmit={form.handleSubmit(onSubmit)}>
          <div className="grid gap-6 md:grid-cols-2">
            <FieldGroup>
              <Field data-invalid={Boolean(form.formState.errors.name) || undefined}>
                <FieldLabel htmlFor="task-name">Name</FieldLabel>
                <Input id="task-name" aria-invalid={Boolean(form.formState.errors.name)} {...form.register("name")} />
                <FieldError errors={[form.formState.errors.name]} />
              </Field>
              <div className="grid gap-5 sm:grid-cols-2">
                <Field>
                  <FieldLabel>Status</FieldLabel>
                  <Controller
                    control={form.control}
                    name="status"
                    render={({ field }) => (
                      <Select value={field.value} onValueChange={(value) => field.onChange(value ?? "BACKLOG")}>
                        <SelectTrigger className="w-full">
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectGroup>
                            {statusOptions.map((status) => (
                              <SelectItem key={status} value={status}>
                                {taskStatusMap[status]?.label ?? status}
                              </SelectItem>
                            ))}
                          </SelectGroup>
                        </SelectContent>
                      </Select>
                    )}
                  />
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
              </div>
              <div className="grid gap-5 sm:grid-cols-2">
                <Field data-invalid={Boolean(form.formState.errors.startDate) || undefined}>
                  <FieldLabel htmlFor="task-start">Start date</FieldLabel>
                  <Input id="task-start" type="date" {...form.register("startDate")} />
                  <FieldError errors={[form.formState.errors.startDate]} />
                </Field>
                <Field data-invalid={Boolean(form.formState.errors.dueDate) || undefined}>
                  <FieldLabel htmlFor="task-due">Due date</FieldLabel>
                  <Input id="task-due" type="date" {...form.register("dueDate")} />
                  <FieldError errors={[form.formState.errors.dueDate]} />
                </Field>
              </div>
              <Field className="min-h-0 flex-1">
                <FieldLabel htmlFor="task-description">Description</FieldLabel>
                <Textarea id="task-description" rows={8} className="min-h-40" {...form.register("description")} />
              </Field>
            </FieldGroup>
            <FieldGroup>
              <Field>
                <FieldLabel>Assignees</FieldLabel>
                <Controller
                  control={form.control}
                  name="assigneeIds"
                  render={({ field }) => (
                    <div className="flex max-h-56 flex-col gap-2 overflow-auto rounded-lg border p-3">
                      {members.length === 0 ? (
                        <p className="text-sm text-muted-foreground">No project members to assign.</p>
                      ) : (
                        members.map((member) => {
                          const checked = field.value.includes(member.userId)
                          return (
                            <Field key={member.userId} orientation="horizontal">
                              <Checkbox
                                checked={checked}
                                onCheckedChange={(value) => {
                                  const next = value
                                    ? [...field.value, member.userId]
                                    : field.value.filter((id) => id !== member.userId)
                                  field.onChange(next)
                                }}
                              />
                              <FieldLabel>{member.name}</FieldLabel>
                            </Field>
                          )
                        })
                      )}
                    </div>
                  )}
                />
              </Field>
              <Field>
                <FieldLabel htmlFor="task-hours">Estimated hours</FieldLabel>
                <Input id="task-hours" type="number" min={0} step="0.5" {...form.register("estimatedHours")} />
              </Field>
              <Field>
                <FieldLabel htmlFor="task-labels">Labels</FieldLabel>
                <Input id="task-labels" placeholder="api, frontend" {...form.register("labels")} />
              </Field>
            </FieldGroup>
          </div>
          {message ? <p className="text-sm text-destructive">{message}</p> : null}
          <DialogFooter className="-mx-4 -mb-4 sm:-mx-6 sm:-mb-6">
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
