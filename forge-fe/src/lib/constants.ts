export const APP_NAME = "Forge"

export const RETURN_TO_KEY = "workspace.returnTo"

export const DEFAULT_PAGE_SIZE = 20

export const ROLE_SUPER_ADMIN = "SUPER_ADMIN"
export const ROLE_ADMIN = "ADMIN"
export const ROLE_PROJECT_MANAGER = "PROJECT_MANAGER"
export const ROLE_RESOURCE_MANAGER = "RESOURCE_MANAGER"
export const ROLE_MEMBER = "MEMBER"

export const AUDIT_PAGE_SIZE = 50

export const NOTIFICATION_TYPES = [
  "TASK_ASSIGNED",
  "TASK_DUE_SOON",
  "TASK_OVERDUE",
  "TASK_STATUS_CHANGED",
  "RESOURCE_OVERLOAD",
  "PROJECT_DUE_SOON",
  "MEMBER_REMOVED",
] as const

export const PERMISSIONS = {
  userManage: "user.manage",
  roleManage: "role.manage",
  departmentManage: "department.manage",
  projectCreate: "project.create",
  projectDelete: "project.delete",
  projectReadAll: "project.read_all",
  taskManage: "task.manage",
  resourceAllocate: "resource.allocate",
  capacityView: "capacity.view",
  reportExport: "report.export",
  auditRead: "audit.read",
  systemConfigure: "system.configure",
} as const
