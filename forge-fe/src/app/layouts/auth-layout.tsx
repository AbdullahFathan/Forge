import { Outlet } from "react-router"

export function AuthLayout() {
  return (
    <div className="flex min-h-svh items-center justify-center p-4 md:p-6">
      <Outlet />
    </div>
  )
}
