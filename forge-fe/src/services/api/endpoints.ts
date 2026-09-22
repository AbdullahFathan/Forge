export const endpoints = {
  login: "/auth/login",
  refresh: "/auth/refresh",
  logout: "/auth/logout",
  me: "/users/me",
  users: "/users",
  user: (id: string) => `/users/${id}`,
  departments: "/departments",
  department: (id: string) => `/departments/${id}`,
  roles: "/roles",
} as const
