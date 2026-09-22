import { describe, expect, it } from "vitest"

import {
  canAddSubtask,
  canArchiveProject,
  nextProjectStatuses,
  nextTaskStatuses,
} from "@/lib/lifecycle"

describe("nextProjectStatuses", () => {
  it("includes current and legal next values", () => {
    expect(nextProjectStatuses("DRAFT")).toEqual(["DRAFT", "ACTIVE", "ON_HOLD"])
    expect(nextProjectStatuses("ACTIVE")).toEqual(["ACTIVE", "ON_HOLD", "COMPLETED"])
    expect(nextProjectStatuses("ARCHIVED")).toEqual(["ARCHIVED"])
  })
})

describe("canArchiveProject", () => {
  it("allows archive from on hold or completed", () => {
    expect(canArchiveProject("ON_HOLD")).toBe(true)
    expect(canArchiveProject("COMPLETED")).toBe(true)
    expect(canArchiveProject("ACTIVE")).toBe(false)
  })
})

describe("nextTaskStatuses", () => {
  it("mirrors the backend machine", () => {
    expect(nextTaskStatuses("BACKLOG")).toEqual(["BACKLOG", "TODO", "BLOCKED"])
    expect(nextTaskStatuses("DONE")).toEqual(["DONE", "IN_REVIEW"])
  })
})

describe("canAddSubtask", () => {
  it("is false for nested tasks", () => {
    expect(canAddSubtask(null)).toBe(true)
    expect(canAddSubtask("parent-id")).toBe(false)
  })
})
