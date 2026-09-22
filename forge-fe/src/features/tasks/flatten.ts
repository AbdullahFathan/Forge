import type { TaskPublic } from "@/types/task"

export function flattenTasks(items: TaskPublic[]) {
  const rows: { task: TaskPublic; depth: 0 | 1 }[] = []
  for (const task of items) {
    rows.push({ task, depth: 0 })
    for (const sub of task.subtasks ?? []) {
      rows.push({ task: sub, depth: 1 })
    }
  }
  return rows
}

export function collectTaskWarnings(items: TaskPublic[]) {
  return flattenTasks(items)
    .map(({ task }) => task)
    .filter((task) => task.warnings.length > 0)
}
