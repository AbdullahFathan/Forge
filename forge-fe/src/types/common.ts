export type Department = {
  id: string
  name: string
  createdAt: string
  updatedAt: string
}

export type RoleSummary = {
  id: string
  code: string
  name: string
  isSystem: boolean
  permissionCodes: string[]
}

export type Permission = {
  id: string
  code: string
  name: string
}
