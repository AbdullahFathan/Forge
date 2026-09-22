export const queryKeys = {
  me: ["users", "me"] as const,
  users: (params: Record<string, unknown>) => ["users", params] as const,
  departments: (params: Record<string, unknown>) => ["departments", params] as const,
  roles: ["roles"] as const,
  projects: (params: Record<string, unknown>) => ["projects", params] as const,
  project: (id: string) => ["projects", id] as const,
  projectMembers: (id: string, params: Record<string, unknown>) =>
    ["projects", id, "members", params] as const,
  projectTasks: (id: string, params: Record<string, unknown>) =>
    ["projects", id, "tasks", params] as const,
  projectActivity: (id: string, params: Record<string, unknown>) =>
    ["projects", id, "activity", params] as const,
  task: (id: string) => ["tasks", id] as const,
  taskComments: (id: string) => ["tasks", id, "comments"] as const,
  allocations: (params: Record<string, unknown>) => ["resources", "allocations", params] as const,
  capacity: (params: Record<string, unknown>) => ["resources", "capacity", params] as const,
  matrix: (params: Record<string, unknown>) => ["resources", "matrix", params] as const,
  availability: (params: Record<string, unknown>) => ["resources", "availability", params] as const,
  overloadAlerts: ["resources", "overload-alerts"] as const,
  workload: (userId?: string) => ["me", "workload", userId ?? "self"] as const,
  dashboard: (kind: string) => ["dashboards", kind] as const,
  notifications: (params: Record<string, unknown>) => ["notifications", params] as const,
  notificationsUnread: ["notifications", "unread-count"] as const,
  auditLogs: (params: Record<string, unknown>) => ["audit-logs", params] as const,
  reports: (kind: string, params: Record<string, unknown>) => ["reports", kind, params] as const,
}
