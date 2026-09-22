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
}
