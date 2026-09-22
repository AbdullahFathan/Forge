export const APP_NAME = "WorkSpace"

export const RETURN_TO_KEY = "workspace.returnTo"

export const DEFAULT_PAGE_SIZE = 20

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
