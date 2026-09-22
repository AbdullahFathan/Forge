import { describe, expect, it } from "vitest"

import { taskFormSchema } from "@/features/tasks/schema"

describe("taskFormSchema", () => {
  const valid = {
    name: "Cutover",
    description: "",
    status: "BACKLOG" as const,
    priority: "MEDIUM" as const,
    estimatedHours: "8",
    startDate: "2026-01-01",
    dueDate: "2026-01-10",
    labels: "ops",
    assigneeIds: ["u1"],
  }

  it("accepts a valid task", () => {
    expect(taskFormSchema.safeParse(valid).success).toBe(true)
  })

  it("blocks due before start", () => {
    const result = taskFormSchema.safeParse({ ...valid, dueDate: "2025-12-01" })
    expect(result.success).toBe(false)
  })
})
