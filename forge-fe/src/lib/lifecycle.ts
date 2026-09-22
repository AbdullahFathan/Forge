export const PROJECT_STATUSES = ["DRAFT", "ACTIVE", "ON_HOLD", "COMPLETED", "ARCHIVED"] as const
export const TASK_STATUSES = ["BACKLOG", "TODO", "IN_PROGRESS", "IN_REVIEW", "DONE", "BLOCKED"] as const
export const PRIORITIES = ["LOW", "MEDIUM", "HIGH", "CRITICAL"] as const
export const PROJECT_ROLES = ["LEAD", "MEMBER", "VIEWER"] as const

const projectTransitions: Record<string, string[]> = {
  DRAFT: ["ACTIVE", "ON_HOLD"],
  ACTIVE: ["ON_HOLD", "COMPLETED"],
  ON_HOLD: ["ACTIVE", "ARCHIVED"],
  COMPLETED: ["ARCHIVED"],
  ARCHIVED: [],
}

const taskTransitions: Record<string, string[]> = {
  BACKLOG: ["TODO", "BLOCKED"],
  TODO: ["BACKLOG", "IN_PROGRESS", "BLOCKED"],
  IN_PROGRESS: ["TODO", "IN_REVIEW", "BLOCKED", "DONE"],
  IN_REVIEW: ["IN_PROGRESS", "DONE", "BLOCKED"],
  DONE: ["IN_REVIEW"],
  BLOCKED: ["BACKLOG", "TODO", "IN_PROGRESS"],
}

export function nextProjectStatuses(current: string) {
  const next = projectTransitions[current] ?? []
  return [current, ...next.filter((status) => status !== current)]
}

export function createProjectStatuses() {
  return ["DRAFT", "ACTIVE", "ON_HOLD"] as const
}

export function canArchiveProject(status: string) {
  return (projectTransitions[status] ?? []).includes("ARCHIVED")
}

export function nextTaskStatuses(current: string) {
  const next = taskTransitions[current] ?? []
  return [current, ...next.filter((status) => status !== current)]
}

export function canAddSubtask(parentTaskId: string | null | undefined) {
  return !parentTaskId
}
