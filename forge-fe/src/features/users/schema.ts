import { z } from "zod"

export const userFormSchema = z.object({
  name: z.string().trim().min(2, "Name is required"),
  email: z.email("Enter a valid email"),
  password: z.string(),
  roleId: z.string().min(1, "Role is required"),
  departmentId: z.string(),
  capacityHoursPerDay: z.number().min(1, "At least 1 hour").max(24, "At most 24 hours"),
  isActive: z.boolean(),
  skills: z.string(),
})

export type UserFormValues = z.infer<typeof userFormSchema>

export function refineUserPassword(values: UserFormValues, isEdit: boolean, ctx: z.RefinementCtx) {
  if (!isEdit && values.password.length < 8) {
    ctx.addIssue({
      code: "custom",
      path: ["password"],
      message: "Use at least 8 characters",
    })
  }
  if (isEdit && values.password.length > 0 && values.password.length < 8) {
    ctx.addIssue({
      code: "custom",
      path: ["password"],
      message: "Use at least 8 characters",
    })
  }
}

export function toUserPayload(values: UserFormValues, isEdit: boolean) {
  const skills = values.skills
    .split(",")
    .map((skill) => skill.trim())
    .filter(Boolean)
  const departmentId = values.departmentId === "none" ? null : values.departmentId
  const shared = {
    name: values.name,
    roleId: values.roleId,
    capacityHoursPerDay: values.capacityHoursPerDay,
    isActive: values.isActive,
    skills,
    ...(values.password ? { password: values.password } : {}),
  }
  if (isEdit) {
    return departmentId
      ? { ...shared, departmentId }
      : { ...shared, clearDepartment: true as const }
  }
  return {
    ...shared,
    email: values.email,
    password: values.password,
    ...(departmentId ? { departmentId } : {}),
  }
}

