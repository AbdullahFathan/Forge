import { EmptyState } from "@/components/common/empty-state"
import { PageHeader } from "@/components/common/page-header"

export function DashboardPage() {
  return (
    <div className="flex flex-col gap-4">
      <PageHeader
        title="Dashboard"
        description="Role-specific charts land in a later phase. Your navigation already follows your permissions."
      />
      <EmptyState
        title="Nothing to review yet"
        description="Project, capacity, and notification summaries will show up here."
      />
    </div>
  )
}
