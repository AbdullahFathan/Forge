import { Link } from "react-router"

import { Button } from "@/components/ui/button"
import { EmptyState } from "@/components/common/empty-state"

export function ForbiddenPage() {
  return (
    <EmptyState
      title="You do not have access"
      description="This page is hidden from your role. Ask an administrator if you need it."
      action={
        <Button render={<Link to="/" />}>
          Back to dashboard
        </Button>
      }
    />
  )
}
