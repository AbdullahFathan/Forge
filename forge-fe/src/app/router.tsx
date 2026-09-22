import { Navigate, Outlet, Route, Routes, useLocation } from "react-router"

import { AppLayout } from "@/app/layouts/app-layout"
import { AuthLayout } from "@/app/layouts/auth-layout"
import { ForbiddenPage } from "@/components/common/forbidden-page"
import { LaterPhasePage } from "@/components/common/later-phase-page"
import { NotFoundPage } from "@/components/common/not-found-page"
import { LoginPage } from "@/features/auth/pages/login-page"
import { DashboardPage } from "@/features/dashboard/pages/dashboard-page"
import { DevPage } from "@/features/dashboard/pages/dev-page"
import { DepartmentsPage } from "@/features/departments/pages/departments-page"
import { ProfilePage } from "@/features/profile/pages/profile-page"
import { UsersPage } from "@/features/users/pages/users-page"
import { can } from "@/lib/auth"
import { PERMISSIONS } from "@/lib/constants"
import { rememberReturnTo } from "@/lib/return-to"
import { useSession } from "@/lib/session"
import { getAccessToken } from "@/lib/storage"

function RequireAuth() {
  const location = useLocation()
  if (!getAccessToken()) {
    rememberReturnTo(`${location.pathname}${location.search}`)
    return <Navigate to="/login" replace />
  }
  return <Outlet />
}

function GuestOnly() {
  if (getAccessToken()) return <Navigate to="/" replace />
  return <Outlet />
}

function RequirePermission({ code }: { code: string }) {
  const permissions = useSession((state) => state.permissions)
  if (!can(permissions, code)) return <ForbiddenPage />
  return <Outlet />
}

export function AppRouter() {
  return (
    <Routes>
      <Route element={<AuthLayout />}>
        <Route element={<GuestOnly />}>
          <Route path="/login" element={<LoginPage />} />
        </Route>
      </Route>
      <Route element={<RequireAuth />}>
        <Route element={<AppLayout />}>
          <Route index element={<DashboardPage />} />
          <Route path="users/me" element={<ProfilePage />} />
          <Route element={<RequirePermission code={PERMISSIONS.userManage} />}>
            <Route path="users" element={<UsersPage />} />
          </Route>
          <Route element={<RequirePermission code={PERMISSIONS.departmentManage} />}>
            <Route path="departments" element={<DepartmentsPage />} />
          </Route>
          <Route path="projects" element={<LaterPhasePage title="Projects" />} />
          <Route path="me/workload" element={<LaterPhasePage title="My workload" />} />
          <Route path="notifications" element={<LaterPhasePage title="Notifications" />} />
          <Route element={<RequirePermission code={PERMISSIONS.capacityView} />}>
            <Route path="resources" element={<LaterPhasePage title="Resources" />} />
          </Route>
          <Route element={<RequirePermission code={PERMISSIONS.reportExport} />}>
            <Route path="reports" element={<LaterPhasePage title="Reports" />} />
          </Route>
          <Route element={<RequirePermission code={PERMISSIONS.auditRead} />}>
            <Route path="audit" element={<LaterPhasePage title="Audit log" />} />
          </Route>
          <Route element={<RequirePermission code={PERMISSIONS.roleManage} />}>
            <Route path="roles" element={<LaterPhasePage title="Roles" />} />
          </Route>
          <Route element={<RequirePermission code={PERMISSIONS.systemConfigure} />}>
            <Route path="holidays" element={<LaterPhasePage title="Holidays" />} />
          </Route>
          {import.meta.env.DEV ? <Route path="dev" element={<DevPage />} /> : null}
        </Route>
      </Route>
      <Route path="*" element={<NotFoundPage />} />
    </Routes>
  )
}
