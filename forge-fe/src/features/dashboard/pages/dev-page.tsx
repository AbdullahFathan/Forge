import { PageHeader } from "@/components/common/page-header"
import { PriorityBadge } from "@/components/common/priority-badge"
import { StatusBadge } from "@/components/common/status-badge"

const projects = ["DRAFT", "ACTIVE", "ON_HOLD", "COMPLETED", "ARCHIVED"]
const tasks = ["BACKLOG", "TODO", "IN_PROGRESS", "IN_REVIEW", "DONE", "BLOCKED"]
const priorities = ["LOW", "MEDIUM", "HIGH", "CRITICAL"]

export function DevPage() {
  return (
    <div className="flex flex-col gap-4">
      <PageHeader title="Components" description="Shared status and priority badges. Remove this route before Phase 5." />
      <div className="flex flex-wrap gap-2">
        {projects.map((value) => (
          <StatusBadge key={value} kind="project" value={value} />
        ))}
      </div>
      <div className="flex flex-wrap gap-2">
        {tasks.map((value) => (
          <StatusBadge key={value} kind="task" value={value} />
        ))}
      </div>
      <div className="flex flex-wrap gap-2">
        {priorities.map((value) => (
          <PriorityBadge key={value} value={value} />
        ))}
      </div>
    </div>
  )
}
