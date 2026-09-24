import {
  PERMISSIONS,
  ROLE_ADMIN,
  ROLE_PROJECT_MANAGER,
  ROLE_RESOURCE_MANAGER,
  ROLE_SUPER_ADMIN,
} from "@/lib/constants"

export type DashboardKind = "executive" | "project-manager" | "resource-manager" | "member"

export function dashboardKind(role: string | null | undefined): DashboardKind {
  if (role === ROLE_SUPER_ADMIN || role === ROLE_ADMIN) return "executive"
  if (role === ROLE_PROJECT_MANAGER) return "project-manager"
  if (role === ROLE_RESOURCE_MANAGER) return "resource-manager"
  return "member"
}

export function can(permissions: readonly string[] | undefined, code: string) {
  return Boolean(permissions?.includes(code))
}

export type NavItem = {
  label: string
  to: string
  permission?: string
  anyPermissions?: readonly string[]
}

export const navItems: NavItem[] = [
  { label: "Dashboard", to: "/" },
  { label: "Projects", to: "/projects" },
  { label: "My workload", to: "/me/workload" },
  {
    label: "Resources",
    to: "/resources",
    anyPermissions: [PERMISSIONS.capacityView, PERMISSIONS.resourceAllocate],
  },
  { label: "Reports", to: "/reports", permission: PERMISSIONS.reportExport },
  { label: "Notifications", to: "/notifications" },
  { label: "Audit log", to: "/audit", permission: PERMISSIONS.auditRead },
  { label: "Users", to: "/users", permission: PERMISSIONS.userManage },
  { label: "Departments", to: "/departments", permission: PERMISSIONS.departmentManage },
  { label: "Roles", to: "/roles", permission: PERMISSIONS.roleManage },
  { label: "Holidays", to: "/holidays", permission: PERMISSIONS.departmentManage },
]

export function visibleNav(permissions: readonly string[]) {
  return navItems.filter((item) => {
    if (item.anyPermissions?.length) {
      return item.anyPermissions.some((code) => can(permissions, code))
    }
    return !item.permission || can(permissions, item.permission)
  })
}

export function formatRole(code: string | null | undefined) {
  if (!code) return "Unknown"
  return code
    .split("_")
    .map((part) => part.charAt(0) + part.slice(1).toLowerCase())
    .join(" ")
}
