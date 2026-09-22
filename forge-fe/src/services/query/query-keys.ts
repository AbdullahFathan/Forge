export const queryKeys = {
  me: ["users", "me"] as const,
  users: (params: Record<string, unknown>) => ["users", params] as const,
  departments: (params: Record<string, unknown>) => ["departments", params] as const,
  roles: ["roles"] as const,
}
