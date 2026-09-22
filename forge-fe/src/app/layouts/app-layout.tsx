import { Outlet } from "react-router"

import { AppSidebar } from "@/components/common/sidebar"
import { Navbar } from "@/components/common/navbar"
import { SidebarInset, SidebarProvider } from "@/components/ui/sidebar"
import { TooltipProvider } from "@/components/ui/tooltip"
import { useCurrentUser } from "@/features/profile/hooks/use-current-user"

export function AppLayout() {
  useCurrentUser()

  return (
    <TooltipProvider>
      <SidebarProvider>
        <AppSidebar />
        <SidebarInset>
          <Navbar />
          <div className="flex flex-1 flex-col gap-4 p-4 md:p-6">
            <Outlet />
          </div>
        </SidebarInset>
      </SidebarProvider>
    </TooltipProvider>
  )
}
