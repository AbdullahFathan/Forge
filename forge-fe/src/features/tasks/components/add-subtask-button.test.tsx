import { describe, expect, it } from "vitest"

import { AddSubtaskButton } from "@/features/tasks/components/add-subtask-button"

describe("AddSubtaskButton", () => {
  it("does not render on a nested task", () => {
    const node = AddSubtaskButton({ parentTaskId: "parent-1", onClick: () => undefined })
    expect(node).toBeNull()
  })

  it("renders on a root task", () => {
    const node = AddSubtaskButton({ parentTaskId: null, onClick: () => undefined })
    expect(node).not.toBeNull()
  })
})
