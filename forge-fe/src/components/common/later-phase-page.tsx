import { EmptyState } from "@/components/common/empty-state"
import { PageHeader } from "@/components/common/page-header"

export function LaterPhasePage({ title }: { title: string }) {
  return (
    <div className="flex flex-col gap-4">
      <PageHeader title={title} description="This screen lands in a later phase." />
      <EmptyState title="Not in this release yet" description="The API is ready. This view is still to come." />
    </div>
  )
}
