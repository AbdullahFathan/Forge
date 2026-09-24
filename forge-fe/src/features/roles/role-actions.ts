export function roleActionsVisible(role: { isSystem: boolean }) {
  return !role.isSystem
}
