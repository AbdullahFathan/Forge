export type AuditAction = "CREATED" | "UPDATED" | "DELETED"

export type AuditLogPublic = {
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

export const AUDIT_ENTITY_TYPES = [
  "Project",
  "Task",
  "User",
  "ProjectMember",
  "ResourceAllocation",
] as const
