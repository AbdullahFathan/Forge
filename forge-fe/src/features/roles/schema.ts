import { z } from "zod"

export const roleFormSchema = z.object({
  name: z.string().trim().min(2, "Name is required").max(128, "Name is too long"),
  permissionCodes: z.array(z.string().min(1)).min(1, "Select at least one permission"),
})

export type RoleFormValues = z.infer<typeof roleFormSchema>
