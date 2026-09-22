import { PERMISSIONS } from "@/lib/constants"

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
  { label: "Holidays", to: "/holidays", permission: PERMISSIONS.systemConfigure },
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
