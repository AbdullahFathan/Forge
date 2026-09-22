import { PERMISSIONS } from "@/lib/constants"
import { can } from "@/lib/auth"
import type { ProjectMember, ProjectPublic } from "@/types/project"
import type { TaskPublic } from "@/types/task"
import type { PublicUser } from "@/types/user"

const ADMIN_ROLES = new Set(["SUPER_ADMIN", "ADMIN"])

export function canManageProject(
  project: Pick<ProjectPublic, "ownerId"> | null | undefined,
  members: ProjectMember[] | undefined,
  user: PublicUser | null | undefined,
  role: string | null | undefined,
) {
  if (!user || !project) return false
  if (role && ADMIN_ROLES.has(role)) return true
  if (project.ownerId === user.id) return true
  return members?.some((member) => member.userId === user.id && member.projectRole === "LEAD") ?? false
}

export function projectRoleFor(
  members: ProjectMember[] | undefined,
  userId: string | undefined,
) {
  if (!userId) return null
  return members?.find((member) => member.userId === userId)?.projectRole ?? null
}

export function isProjectViewer(role: string | null | undefined) {
  return role === "VIEWER"
}

export function canWriteTasks(
  project: Pick<ProjectPublic, "ownerId"> | null | undefined,
  members: ProjectMember[] | undefined,
  user: PublicUser | null | undefined,
  role: string | null | undefined,
  permissions: readonly string[] | undefined,
) {
  if (!can(permissions, PERMISSIONS.taskManage)) return false
  if (isProjectViewer(projectRoleFor(members, user?.id))) return false
  return canManageProject(project, members, user, role) || Boolean(user)
}

export function canMutateTaskFields(
  task: TaskPublic,
  project: Pick<ProjectPublic, "ownerId"> | null | undefined,
  members: ProjectMember[] | undefined,
  user: PublicUser | null | undefined,
  role: string | null | undefined,
  permissions: readonly string[] | undefined,
) {
  if (!can(permissions, PERMISSIONS.taskManage)) return false
  if (isProjectViewer(projectRoleFor(members, user?.id))) return false
  if (canManageProject(project, members, user, role)) return true
  return Boolean(user && task.assignees.some((assignee) => assignee.id === user.id))
}

export function canAssigneeOnlyPatch(
  task: TaskPublic,
  project: Pick<ProjectPublic, "ownerId"> | null | undefined,
  members: ProjectMember[] | undefined,
  user: PublicUser | null | undefined,
  role: string | null | undefined,
) {
  if (canManageProject(project, members, user, role)) return false
  return Boolean(user && task.assignees.some((assignee) => assignee.id === user.id))
}
