export type ProjectStatus = "DRAFT" | "ACTIVE" | "ON_HOLD" | "COMPLETED" | "ARCHIVED"

export type ProjectPriority = "LOW" | "MEDIUM" | "HIGH" | "CRITICAL"

export type ProjectRole = "LEAD" | "MEMBER" | "VIEWER"

export type TaskCounts = {
  BACKLOG: number
  TODO: number
  IN_PROGRESS: number
  IN_REVIEW: number
  DONE: number
  BLOCKED: number
  total: number
}

export type ProjectPublic = {
  id: string
  name: string
  description: string
  status: ProjectStatus | string
  priority: ProjectPriority | string
  startDate: string
  targetEndDate: string
  ownerId: string
  ownerName?: string
  departmentId: string | null
  departmentName?: string | null
  tags: string[]
  completionPercent: number
  createdAt: string
  updatedAt: string
}

export type ProjectSummary = ProjectPublic & {
  taskCounts: TaskCounts
  memberCount: number
  activity: ProjectActivity[]
}

export type ProjectMember = {
  userId: string
  name: string
  email?: string
  departmentId: string | null
  departmentName?: string | null
  projectRole: ProjectRole | string
}

export type ProjectActivity = {
  id: string
  userId: string
  ipAddress: string
  entityType: string
  entityId: string
  projectId?: string | null
  action: string
  before?: unknown
  after?: unknown
  createdAt: string
}

export const ARCHIVE_PROJECT_COPY = "Archive project"
