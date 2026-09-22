import { z } from "zod"

import type { AllocationPatch, AllocationWrite } from "@/types/resource"

const dateRe = /^\d{4}-\d{2}-\d{2}$/

export const ALLOCATION_ROLES = ["LEAD", "MEMBER", "VIEWER"] as const

export const allocationFormSchema = z
  .object({
    userId: z.string().min(1, "User is required"),
    projectId: z.string().min(1, "Project is required"),
    allocationPercent: z.number().min(0, "Minimum 0%").max(100, "Maximum 100%"),
    startDate: z.string().regex(dateRe, "Use YYYY-MM-DD"),
    endDate: z.string().regex(dateRe, "Use YYYY-MM-DD"),
    role: z.enum(ALLOCATION_ROLES),
  })
  .superRefine((values, ctx) => {
    if (values.endDate < values.startDate) {
      ctx.addIssue({
        code: "custom",
        path: ["endDate"],
        message: "End date cannot be before start date",
      })
    }
  })

export type AllocationFormValues = z.infer<typeof allocationFormSchema>

export function toAllocationCreate(values: AllocationFormValues): AllocationWrite {
  return {
    userId: values.userId,
    projectId: values.projectId,
    allocationPercent: values.allocationPercent,
    startDate: values.startDate,
    endDate: values.endDate,
    role: values.role,
  }
}

export function toAllocationPatch(values: AllocationFormValues): AllocationPatch {
  return {
    allocationPercent: values.allocationPercent,
    startDate: values.startDate,
    endDate: values.endDate,
    role: values.role,
  }
}

export const capacityFilterSchema = z.object({
  from: z.string().regex(dateRe),
  to: z.string().regex(dateRe),
  granularity: z.enum(["week", "month"]),
  departmentId: z.string(),
  skill: z.string(),
})
