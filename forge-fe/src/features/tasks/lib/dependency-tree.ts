import { flattenTasks } from "@/features/tasks/flatten"
import type { TaskDepRef, TaskPublic } from "@/types/task"

export type HydratedDep = {
  id: string
  taskId: string
  name: string
  status?: string
}

export type DependencyTreeModel = {
  predecessors: HydratedDep[]
  current: { id: string; name: string; status: string }
  dependents: HydratedDep[]
}

function shortId(taskId: string) {
  return taskId.slice(0, 8)
}

export function taskIndex(allTasks: TaskPublic[]) {
  const byId = new Map<string, TaskPublic>()
  for (const { task } of flattenTasks(allTasks)) {
    byId.set(task.id, task)
  }
  return byId
}

export function hydrateDepRef(ref: TaskDepRef, byId: Map<string, TaskPublic>): HydratedDep {
  const listed = byId.get(ref.taskId)
  return {
    id: ref.id,
    taskId: ref.taskId,
    name: ref.name || listed?.name || shortId(ref.taskId),
    status: ref.status || listed?.status,
  }
}

export function dependencyTreeModel(task: TaskPublic, allTasks: TaskPublic[]): DependencyTreeModel {
  const byId = taskIndex(allTasks)
  return {
    predecessors: task.dependsOn.map((ref) => hydrateDepRef(ref, byId)),
    current: { id: task.id, name: task.name, status: String(task.status) },
    dependents: task.dependents.map((ref) => hydrateDepRef(ref, byId)),
  }
}
