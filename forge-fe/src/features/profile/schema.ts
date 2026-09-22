import { z } from "zod"

export const profileSchema = z
  .object({
    name: z.string().trim().min(2, "Name is required"),
    currentPassword: z.string(),
    password: z.string(),
    emailNotificationsEnabled: z.boolean(),
  })
  .superRefine((values, ctx) => {
    const changing = values.password.length > 0 || values.currentPassword.length > 0
    if (!changing) return
    if (values.currentPassword.length < 1) {
      ctx.addIssue({ code: "custom", path: ["currentPassword"], message: "Enter your current password" })
    }
    if (values.password.length < 8) {
      ctx.addIssue({ code: "custom", path: ["password"], message: "Use at least 8 characters" })
    }
  })

export type ProfileValues = z.infer<typeof profileSchema>
