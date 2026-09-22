import { z } from "zod"

import { createProjectStatuses, PRIORITIES } from "@/lib/lifecycle"
import type { ProjectWrite } from "@/features/projects/api/projects"

const dateRe = /^\d{4}-\d{2}-\d{2}$/

export const projectFormSchema = z
  .object({
    name: z.string().trim().min(3, "Name must be at least 3 characters"),
    description: z.string(),
    status: z.string().min(1),
    priority: z.enum(PRIORITIES),
    startDate: z.string().regex(dateRe, "Use YYYY-MM-DD"),
    targetEndDate: z.string().regex(dateRe, "Use YYYY-MM-DD"),
    departmentId: z.string(),
    tags: z.string(),
  })
  .superRefine((values, ctx) => {
    if (values.targetEndDate < values.startDate) {
      ctx.addIssue({
        code: "custom",
        path: ["targetEndDate"],
        message: "End date cannot be before start date",
      })
    }
  })

export type ProjectFormValues = z.infer<typeof projectFormSchema>

export function parseCsvList(value: string) {
  return value
    .split(",")
    .map((item) => item.trim())
    .filter(Boolean)
}

export function toProjectPayload(values: ProjectFormValues, ownerId: string, isEdit: boolean): ProjectWrite | Partial<ProjectWrite> {
  const tags = parseCsvList(values.tags)
  const departmentId = values.departmentId === "none" ? null : values.departmentId
  const shared = {
    name: values.name,
    description: values.description,
    status: values.status,
    priority: values.priority,
    startDate: values.startDate,
    targetEndDate: values.targetEndDate,
    tags,
  }
  if (isEdit) {
    return departmentId
      ? { ...shared, departmentId }
      : { ...shared, clearDepartment: true as const }
  }
  return {
    ...shared,
    ownerId,
    ...(departmentId ? { departmentId } : {}),
  }
}

export const createStatusOptions = createProjectStatuses()
