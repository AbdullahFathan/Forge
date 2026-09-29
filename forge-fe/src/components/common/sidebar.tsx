import {
  Bell,
  Building2,
  CalendarDays,
  FileBarChart,
  FolderKanban,
  Gauge,
  LayoutDashboard,
  ListChecks,
  ScrollText,
  Shield,
  Users,
  type LucideIcon,
} from "lucide-react"
import { NavLink, useLocation } from "react-router"

import { APP_NAME } from "@/lib/constants"
import { visibleNav, type NavItem } from "@/lib/auth"
import { useSession } from "@/lib/session"
import {
  Sidebar,
  SidebarContent,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarRail,
  useSidebar,
} from "@/components/ui/sidebar"

const icons: Record<string, LucideIcon> = {
  "/": LayoutDashboard,
  "/projects": FolderKanban,
  "/me/workload": ListChecks,
  "/resources": Gauge,
  "/reports": FileBarChart,
  "/notifications": Bell,
  "/audit": ScrollText,
  "/users": Users,
  "/departments": Building2,
  "/roles": Shield,
  "/holidays": CalendarDays,
}

const adminPaths = new Set(["/audit", "/users", "/departments", "/roles", "/holidays"])

function isItemActive(pathname: string, to: string) {
  if (to === "/" || to === "/users") return pathname === to
  return pathname === to || pathname.startsWith(`${to}/`)
}

function NavItems({ items, pathname }: { items: NavItem[]; pathname: string }) {
  const { setOpenMobile } = useSidebar()
  return (
    <SidebarMenu>
      {items.map((item) => {
        const Icon = icons[item.to]
        const active = isItemActive(pathname, item.to)
        return (
          <SidebarMenuItem key={item.to}>
            <SidebarMenuButton
              render={
                <NavLink
                  to={item.to}
                  end={item.to === "/" || item.to === "/users"}
                  onClick={() => setOpenMobile(false)}
                />
              }
              isActive={active}
            >
              {Icon ? (
                <Icon className={active ? "text-primary" : undefined} aria-hidden />
              ) : null}
              <span>{item.label}</span>
            </SidebarMenuButton>
          </SidebarMenuItem>
        )
      })}
    </SidebarMenu>
  )
}

export function AppSidebar() {
  const permissions = useSession((state) => state.permissions)
  const { pathname } = useLocation()
  const items = visibleNav(permissions)
  const work = items.filter((item) => !adminPaths.has(item.to))
  const admin = items.filter((item) => adminPaths.has(item.to))

  return (
    <Sidebar collapsible="offcanvas">
      <SidebarHeader className="px-4 py-3">
        <span className="text-sm font-semibold tracking-tight">{APP_NAME}</span>
      </SidebarHeader>
      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupContent>
            <NavItems items={work} pathname={pathname} />
          </SidebarGroupContent>
        </SidebarGroup>
        {admin.length > 0 ? (
          <SidebarGroup className="pt-1">
            <SidebarGroupLabel>Admin</SidebarGroupLabel>
            <SidebarGroupContent>
              <NavItems items={admin} pathname={pathname} />
            </SidebarGroupContent>
          </SidebarGroup>
        ) : null}
      </SidebarContent>
      <SidebarRail />
    </Sidebar>
  )
}
