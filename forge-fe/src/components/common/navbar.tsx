import { useQuery } from "@tanstack/react-query"
import { BellIcon } from "lucide-react"
import { Link, useLocation } from "react-router"

import { getUnreadCount } from "@/features/notifications/api/notifications"
import { queryKeys } from "@/services/query/query-keys"

import { formatRole } from "@/lib/auth"
import { usePageCrumb } from "@/lib/breadcrumb"
import { useAuth } from "@/features/auth/hooks/use-auth"
import { Avatar, AvatarFallback } from "@/components/ui/avatar"
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb"
import { Button } from "@/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { Separator } from "@/components/ui/separator"
import { SidebarTrigger } from "@/components/ui/sidebar"

const crumbs: Record<string, string> = {
  "/": "Dashboard",
  "/projects": "Projects",
  "/me/workload": "My workload",
  "/resources": "Resources",
  "/reports": "Reports",
  "/notifications": "Notifications",
  "/audit": "Audit log",
  "/users": "Users",
  "/users/me": "Profile",
  "/departments": "Departments",
  "/roles": "Roles",
  "/holidays": "Holidays",
  "/dev": "Components",
}

function initials(name: string) {
  return name
    .split(" ")
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part.charAt(0).toUpperCase())
    .join("")
}

export function Navbar() {
  const { pathname } = useLocation()
  const { user, role, logout } = useAuth()
  const unread = useQuery({
    queryKey: queryKeys.notificationsUnread,
    queryFn: getUnreadCount,
    refetchInterval: 60_000,
    refetchOnWindowFocus: true,
  })
  const unreadCount = unread.data?.unreadCount ?? 0
  const detailLabel = usePageCrumb((state) => state.detailLabel)
  const projectMatch = pathname.match(/^\/projects\/([^/]+)$/)
  const current = crumbs[pathname] ?? (projectMatch ? "Projects" : "WorkSpace")

  return (
    <header className="flex h-14 shrink-0 items-center gap-2 border-b px-4">
      <SidebarTrigger />
      <Separator orientation="vertical" className="h-4" />
      <Breadcrumb>
        <BreadcrumbList>
          <BreadcrumbItem>
            <BreadcrumbLink render={<Link to="/" />}>WorkSpace</BreadcrumbLink>
          </BreadcrumbItem>
          <BreadcrumbSeparator />
          {projectMatch ? (
            <>
              <BreadcrumbItem>
                <BreadcrumbLink render={<Link to="/projects" />}>Projects</BreadcrumbLink>
              </BreadcrumbItem>
              <BreadcrumbSeparator />
              <BreadcrumbItem>
                <BreadcrumbPage>{detailLabel ?? "Project"}</BreadcrumbPage>
              </BreadcrumbItem>
            </>
          ) : (
            <BreadcrumbItem>
              <BreadcrumbPage>{current}</BreadcrumbPage>
            </BreadcrumbItem>
          )}
        </BreadcrumbList>
      </Breadcrumb>
      <div className="ml-auto flex items-center gap-2">
        <Button
          variant="ghost"
          size="sm"
          render={<Link to="/notifications" />}
          aria-label={`Notifications, ${unreadCount} unread`}
        >
          <BellIcon data-icon="inline-start" />
          <span className="tabular-nums">{unreadCount}</span>
        </Button>
        <DropdownMenu>
          <DropdownMenuTrigger
            render={
              <Button variant="ghost" size="sm" aria-label="Account menu">
                <Avatar className="size-6">
                  <AvatarFallback>{user ? initials(user.name) : "?"}</AvatarFallback>
                </Avatar>
                <span className="max-w-32 truncate">{user?.name ?? "Account"}</span>
              </Button>
            }
          />
          <DropdownMenuContent align="end" className="w-56">
            <DropdownMenuGroup>
              <DropdownMenuLabel className="flex flex-col gap-0.5">
                <span>{user?.name}</span>
                <span className="text-xs font-normal text-muted-foreground">
                  {formatRole(role)}
                  {user?.departmentName ? ` · ${user.departmentName}` : ""}
                </span>
              </DropdownMenuLabel>
            </DropdownMenuGroup>
            <DropdownMenuSeparator />
            <DropdownMenuGroup>
              <DropdownMenuItem render={<Link to="/users/me" />}>Profile</DropdownMenuItem>
              <DropdownMenuItem onClick={() => logout()}>Log out</DropdownMenuItem>
            </DropdownMenuGroup>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </header>
  )
}
