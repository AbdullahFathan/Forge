import { z } from "zod"

export const departmentSchema = z.object({
  name: z.string().trim().min(2, "Name is required"),
})

export type DepartmentValues = z.infer<typeof departmentSchema>
