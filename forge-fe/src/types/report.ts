export type ReportKind = "project-status" | "resource-utilization" | "task-completion"

export type ProjectStatusRow = {
  id: string
  name: string
  status: string
  completionPercent: number
  targetEndDate: string
  ownerId: string
  ownerName?: string
}

export type UtilizationRow = {
  userId: string
  name: string
  utilizationPercent: number
  band: string
}

export type TaskCompletionRow = {
  id: string
  name: string
  done: number
  total: number
  percent: number
}

export type ReportFilters = {
  from?: string
  to?: string
  departmentId?: string
  projectId?: string
  groupBy?: "project" | "user"
}
