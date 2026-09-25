import { Link } from "react-router"

import { Button } from "@/components/ui/button"
import { APP_NAME } from "@/lib/constants"
import { EmptyState } from "@/components/common/empty-state"

export function NotFoundPage() {
  return (
    <div className="flex min-h-svh items-center justify-center p-6">
      <EmptyState
        title="Page not found"
        description={`That address is not part of ${APP_NAME}.`}
        action={
          <Button render={<Link to="/" />}>
            Back to dashboard
          </Button>
        }
      />
    </div>
  )
}
