import { describe, expect, it } from "vitest"

import { TaskDependencyTree } from "@/features/tasks/components/task-dependency-tree"
import type { TaskPublic } from "@/types/task"

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

describe("TaskDependencyTree", () => {
  it("renders two predecessors and a dependent instead of a flat Depends on list", () => {
    const current = task({
      id: "b",
      name: "Implement",
      dependsOn: [
        { id: "d1", taskId: "a", name: "Design", status: "DONE" },
        { id: "d2", taskId: "c", name: "Spec", status: "TODO" },
      ],
      dependents: [{ id: "d3", taskId: "e" }],
    })
    const node = TaskDependencyTree({
      task: current,
      allTasks: [task({ id: "a", name: "Design" }), current, task({ id: "c", name: "Spec" }), task({ id: "e", name: "Review" })],
    })
    const text = JSON.stringify(node)
    expect(text).toContain("Design")
    expect(text).toContain("Spec")
    expect(text).toContain("This task")
    expect(text).toContain("Review")
    expect(text).toContain("Waiting on this task")
    expect(text).not.toContain("Depends on")
  })
})
