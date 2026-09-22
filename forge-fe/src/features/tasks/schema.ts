import { z } from "zod"

import { PRIORITIES, TASK_STATUSES } from "@/lib/lifecycle"
import { parseCsvList } from "@/features/projects/schema"
import type { TaskPatch, TaskWrite } from "@/features/tasks/api/tasks"

const dateRe = /^\d{4}-\d{2}-\d{2}$/
const optionalDate = z.string().refine((value) => value === "" || dateRe.test(value), "Use YYYY-MM-DD")

export const taskFormSchema = z
  .object({
    name: z.string().trim().min(2, "Name is required"),
    description: z.string(),
    status: z.enum(TASK_STATUSES),
    priority: z.enum(PRIORITIES),
    estimatedHours: z.string(),
    startDate: optionalDate,
    dueDate: optionalDate,
    labels: z.string(),
    assigneeIds: z.array(z.string()),
  })
  .superRefine((values, ctx) => {
    if (values.startDate && values.dueDate && values.dueDate < values.startDate) {
      ctx.addIssue({
        code: "custom",
        path: ["dueDate"],
        message: "Due date cannot be before start date",
      })
    }
  })

export type TaskFormValues = z.infer<typeof taskFormSchema>

export const commentSchema = z.object({
  body: z.string().trim().min(1, "Comment is required"),
})

export type CommentValues = z.infer<typeof commentSchema>

export function toTaskCreate(values: TaskFormValues, parentTaskId?: string | null): TaskWrite {
  const hours = values.estimatedHours.trim() === "" ? undefined : Number(values.estimatedHours)
  return {
    name: values.name,
    description: values.description,
    status: values.status,
    priority: values.priority,
    labels: parseCsvList(values.labels),
    assigneeIds: values.assigneeIds,
    ...(parentTaskId ? { parentTaskId } : {}),
    ...(Number.isFinite(hours) ? { estimatedHours: hours } : {}),
    ...(values.startDate ? { startDate: values.startDate } : {}),
    ...(values.dueDate ? { dueDate: values.dueDate } : {}),
  }
}

export function toTaskPatch(values: TaskFormValues): TaskPatch {
  const hours = values.estimatedHours.trim() === "" ? null : Number(values.estimatedHours)
  return {
    name: values.name,
    description: values.description,
    status: values.status,
    priority: values.priority,
    labels: parseCsvList(values.labels),
    assigneeIds: values.assigneeIds,
    ...(hours === null
      ? { clearEstimatedHours: true }
      : Number.isFinite(hours)
        ? { estimatedHours: hours }
        : {}),
    ...(values.startDate ? { startDate: values.startDate } : { clearStartDate: true }),
    ...(values.dueDate ? { dueDate: values.dueDate } : { clearDueDate: true }),
  }
}
