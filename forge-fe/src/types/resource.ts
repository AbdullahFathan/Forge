export type AllocationRole = "LEAD" | "MEMBER" | "VIEWER"

export type CapacityBand = "GREEN" | "YELLOW" | "RED"

export type AllocationUserBrief = {
  id: string
  name: string
  departmentId?: string | null
  departmentName?: string | null
}

export type AllocationPublic = {
  id: string
  userId: string
  user: AllocationUserBrief
  projectId: string
  projectName?: string
  allocationPercent: number
  startDate: string
  endDate: string
  role: AllocationRole | string
  warnings: string[]
  overAllocated: boolean
  createdAt: string
  updatedAt: string
}

export type AllocationWrite = {
  userId: string
  projectId: string
  allocationPercent: number
  startDate: string
  endDate: string
  role?: AllocationRole
}

export type AllocationPatch = {
  allocationPercent?: number
  startDate?: string
  endDate?: string
  role?: AllocationRole
}

export type AllocationBreakdown = {
  projectId: string
  projectName?: string
  allocationPercent: number
}

export type CapacityBucket = {
  periodKey: string
  periodStart: string
  periodEnd: string
  utilizationPercent: number
  band: CapacityBand | string
  allocatedHours: number
  effectiveHours: number
  allocationBreakdown: AllocationBreakdown[]
}

export type UserForecast = {
  userId: string
  name: string
  departmentId?: string
  capacityHoursPerDay: number
  skills: string[]
  buckets: CapacityBucket[]
}

export type CapacityForecast = {
  from: string
  to: string
  granularity: "week" | "month" | string
  users: UserForecast[]
}

export type MatrixPeriod = {
  key: string
  start: string
  end: string
}

export type MatrixCell = {
  utilizationPercent: number
  band: CapacityBand | string
}

export type AllocationMatrix = {
  users: { id: string; name: string }[]
  periods: MatrixPeriod[]
  cells: Record<string, Record<string, MatrixCell>>
}

export type AvailabilityItem = {
  userId: string
  name: string
  departmentId?: string
  utilizationPercent: number
  remainingPercent: number
  band: CapacityBand | string
}

export type OverloadAlert = {
  userId: string
  name: string
  maxPercent: number
  from: string
  to: string
}

export type WorkloadProject = {
  projectId: string
  projectName?: string
  allocationPercent: number
  startDate: string
  endDate: string
  role: string
}

export type WorkloadTask = {
  id: string
  projectId: string
  name: string
  status: string
  priority: string
  dueDate: string | null
}

export type Workload = {
  userId: string
  projects: WorkloadProject[]
  tasks: WorkloadTask[]
  weekSeries: CapacityBucket[]
}

export const OVER_ALLOCATED_WARNING = "OVER_ALLOCATED"
