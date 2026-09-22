export type TaskStatus =
  | "BACKLOG"
  | "TODO"
  | "IN_PROGRESS"
  | "IN_REVIEW"
  | "DONE"
  | "BLOCKED"

export type TaskUserBrief = {
  id: string
  name: string
}

export type TaskDepRef = {
  id: string
  taskId: string
  name?: string
  status?: string
}

export type TaskPublic = {
  id: string
  projectId: string
  parentTaskId: string | null
  name: string
  description: string
  status: TaskStatus | string
  priority: string
  estimatedHours: number | null
  startDate: string | null
  dueDate: string | null
  labels: string[]
  position: number
  assignees: TaskUserBrief[]
  dependsOn: TaskDepRef[]
  dependents: TaskDepRef[]
  warnings: string[]
  subtasks?: TaskPublic[]
  createdAt: string
  updatedAt: string
}

export type TaskComment = {
  id: string
  taskId: string
  userId: string
  userName?: string
  body: string
  createdAt: string
}
