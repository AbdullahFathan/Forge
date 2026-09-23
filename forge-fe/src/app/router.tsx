import { Navigate, Outlet, Route, Routes, useLocation } from "react-router"

import { AppLayout } from "@/app/layouts/app-layout"
import { AuthLayout } from "@/app/layouts/auth-layout"
import { ForbiddenPage } from "@/components/common/forbidden-page"
import { LaterPhasePage } from "@/components/common/later-phase-page"
import { NotFoundPage } from "@/components/common/not-found-page"
import { AuditPage } from "@/features/audit/pages/audit-page"
import { NotificationsPage } from "@/features/notifications/pages/notifications-page"
import { ReportsPage } from "@/features/reports/pages/reports-page"
import { LoginPage } from "@/features/auth/pages/login-page"
import { DashboardPage } from "@/features/dashboard/pages/dashboard-page"
import { DevPage } from "@/features/dashboard/pages/dev-page"
import { DepartmentsPage } from "@/features/departments/pages/departments-page"
import { HolidaysPage } from "@/features/holidays/pages/holidays-page"
import { ProfilePage } from "@/features/profile/pages/profile-page"
import { ProjectDetailPage } from "@/features/projects/pages/project-detail-page"
import { ProjectsPage } from "@/features/projects/pages/projects-page"
import { ResourcesPage } from "@/features/resources/pages/resources-page"
import { WorkloadPage } from "@/features/resources/pages/workload-page"
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

function RequireAnyPermission({ codes }: { codes: string[] }) {
  const permissions = useSession((state) => state.permissions)
  if (!codes.some((code) => can(permissions, code))) return <ForbiddenPage />
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
          <Route path="projects" element={<ProjectsPage />} />
          <Route path="projects/:id" element={<ProjectDetailPage />} />
          <Route path="me/workload" element={<WorkloadPage />} />
          <Route path="notifications" element={<NotificationsPage />} />
          <Route
            element={
              <RequireAnyPermission
                codes={[PERMISSIONS.capacityView, PERMISSIONS.resourceAllocate]}
              />
            }
          >
            <Route path="resources" element={<ResourcesPage />} />
          </Route>
          <Route element={<RequirePermission code={PERMISSIONS.reportExport} />}>
            <Route path="reports" element={<ReportsPage />} />
          </Route>
          <Route element={<RequirePermission code={PERMISSIONS.auditRead} />}>
            <Route path="audit" element={<AuditPage />} />
          </Route>
          <Route element={<RequirePermission code={PERMISSIONS.roleManage} />}>
            <Route path="roles" element={<LaterPhasePage title="Roles" />} />
          </Route>
          <Route element={<RequirePermission code={PERMISSIONS.departmentManage} />}>
            <Route path="holidays" element={<HolidaysPage />} />
          </Route>
          {import.meta.env.DEV ? <Route path="dev" element={<DevPage />} /> : null}
        </Route>
      </Route>
      <Route path="*" element={<NotFoundPage />} />
    </Routes>
  )
}
