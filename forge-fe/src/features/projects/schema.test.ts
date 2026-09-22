import { describe, expect, it } from "vitest"

import { projectFormSchema } from "@/features/projects/schema"
import { ARCHIVE_PROJECT_COPY } from "@/types/project"

describe("projectFormSchema", () => {
  const valid = {
    name: "Network rollout",
    description: "",
    status: "DRAFT",
    priority: "HIGH" as const,
    startDate: "2026-01-01",
    targetEndDate: "2026-02-01",
    departmentId: "none",
    tags: "infra",
  }

  it("accepts a valid project", () => {
    expect(projectFormSchema.safeParse(valid).success).toBe(true)
  })

  it("blocks end before start", () => {
    const result = projectFormSchema.safeParse({ ...valid, targetEndDate: "2025-12-01" })
    expect(result.success).toBe(false)
  })

  it("requires a name", () => {
    const result = projectFormSchema.safeParse({ ...valid, name: "ab" })
    expect(result.success).toBe(false)
  })
})

describe("archive copy", () => {
  it("uses Archive project not Delete", () => {
    expect(ARCHIVE_PROJECT_COPY).toBe("Archive project")
    expect(ARCHIVE_PROJECT_COPY.toLowerCase()).not.toContain("delete")
  })
})
