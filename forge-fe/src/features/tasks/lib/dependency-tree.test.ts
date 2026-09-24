import { describe, expect, it } from "vitest"

import { dependencyTreeModel, hydrateDepRef } from "@/features/tasks/lib/dependency-tree"
import type { TaskDepRef, TaskPublic } from "@/types/task"

function task(partial: Partial<TaskPublic> & Pick<TaskPublic, "id" | "name">): TaskPublic {
  return {
    projectId: "p1",
    parentTaskId: null,
    description: "",
    status: "TODO",
    priority: "MEDIUM",
    estimatedHours: null,
    startDate: null,
    dueDate: null,
    labels: [],
    position: 0,
    assignees: [],
    dependsOn: [],
    dependents: [],
    warnings: [],
    createdAt: "",
    updatedAt: "",
    ...partial,
  }
}

describe("hydrateDepRef", () => {
  it("fills name from the project task list when the API omits it", () => {
    const byId = new Map([["task-b", task({ id: "task-b", name: "Write tests", status: "BACKLOG" })]])
    const ref: TaskDepRef = { id: "dep-1", taskId: "task-b" }
    expect(hydrateDepRef(ref, byId)).toEqual({
      id: "dep-1",
      taskId: "task-b",
      name: "Write tests",
      status: "BACKLOG",
    })
  })
})

describe("dependencyTreeModel", () => {
  it("lists two predecessors above the current task", () => {
    const a = task({ id: "a", name: "Design", status: "DONE" })
    const c = task({ id: "c", name: "Spec", status: "IN_PROGRESS" })
    const b = task({
      id: "b",
      name: "Implement",
      status: "TODO",
      dependsOn: [
        { id: "d1", taskId: "a", name: "Design", status: "DONE" },
        { id: "d2", taskId: "c", name: "Spec", status: "IN_PROGRESS" },
      ],
      dependents: [{ id: "d3", taskId: "e" }],
    })
    const e = task({ id: "e", name: "Review", status: "BACKLOG" })
    const model = dependencyTreeModel(b, [a, b, c, e])
    expect(model.predecessors).toHaveLength(2)
    expect(model.predecessors.map((row) => row.name)).toEqual(["Design", "Spec"])
    expect(model.current).toEqual({ id: "b", name: "Implement", status: "TODO" })
    expect(model.dependents).toEqual([
      { id: "d3", taskId: "e", name: "Review", status: "BACKLOG" },
    ])
  })
})
