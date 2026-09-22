import type { AuditLogPublic } from "@/types/audit"
import type { NotificationPublic } from "@/types/notification"
import type { ProjectPublic } from "@/types/project"
import type { AvailabilityItem, Workload } from "@/types/resource"
import type { TaskPublic } from "@/types/task"

export type MonthPoint = {
  month: string
  count: number
}

export type ExecutiveDashboard = {
  activeProjects: number
  completedProjects: number
  lateProjects: number
  orgUtilization: number
  dueSoon: ProjectPublic[]
  topUtilization: AvailabilityItem[]
  monthlyCompleted: MonthPoint[]
}

export type ProjectManagerDashboard = {
  projects: ProjectPublic[]
  blocked: TaskPublic[]
  overdue: TaskPublic[]
  unassigned: TaskPublic[]
  teamCapacity: AvailabilityItem[]
  recentActivity: AuditLogPublic[]
}

export type MemberDashboard = {
  tasks: TaskPublic[]
  weekLoad: Workload
  notifications: NotificationPublic[]
}
