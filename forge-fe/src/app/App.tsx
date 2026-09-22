import { BrowserRouter } from "react-router"

import { AppRouter } from "@/app/router"
import { QueryProvider, SessionProvider, ThemeProvider } from "@/app/providers"
import { Toaster } from "@/components/ui/sonner"

export default function App() {
  return (
    <ThemeProvider>
      <QueryProvider>
        <SessionProvider>
          <BrowserRouter>
            <AppRouter />
          </BrowserRouter>
          <Toaster />
        </SessionProvider>
      </QueryProvider>
    </ThemeProvider>
  )
}
