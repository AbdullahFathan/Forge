import { NavLink, useLocation } from "react-router"

import { APP_NAME } from "@/lib/constants"
import { visibleNav } from "@/lib/auth"
import { useSession } from "@/lib/session"
import {
  Sidebar,
  SidebarContent,
  SidebarGroup,
  SidebarGroupContent,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarRail,
} from "@/components/ui/sidebar"

export function AppSidebar() {
  const permissions = useSession((state) => state.permissions)
  const { pathname } = useLocation()
  const items = visibleNav(permissions)

  return (
    <Sidebar collapsible="offcanvas">
      <SidebarHeader className="px-3 py-3">
        <span className="text-sm font-semibold tracking-tight">{APP_NAME}</span>
      </SidebarHeader>
      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupContent>
            <SidebarMenu>
              {items.map((item) => (
                <SidebarMenuItem key={item.to}>
                  <SidebarMenuButton
                    render={<NavLink to={item.to} end />}
                    isActive={pathname === item.to}
                  >
                    {item.label}
                  </SidebarMenuButton>
                </SidebarMenuItem>
              ))}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>
      <SidebarRail />
    </Sidebar>
  )
}
