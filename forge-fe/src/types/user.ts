export type PublicUser = {
  id: string
  name: string
  email: string
  roleId: string
  roleCode: string
  departmentId: string | null
  departmentName?: string | null
  capacityHoursPerDay: number
  isActive: boolean
  emailNotificationsEnabled: boolean
  skills: string[]
  createdAt: string
  updatedAt: string
}

export type LoginResult = {
  accessToken: string
  expiresIn: number
  user: PublicUser
}
