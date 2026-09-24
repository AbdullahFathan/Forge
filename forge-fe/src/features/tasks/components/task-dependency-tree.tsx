import type { ReactNode } from "react"

import { StatusBadge } from "@/components/common/status-badge"
import { dependencyTreeModel } from "@/features/tasks/lib/dependency-tree"
import type { TaskPublic } from "@/types/task"

export function TaskDependencyTree({
  task,
  allTasks,
  predecessorAction,
}: {
  task: TaskPublic
  allTasks: TaskPublic[]
  predecessorAction?: (depId: string) => ReactNode
}) {
  const model = dependencyTreeModel(task, allTasks)

  return (
    <div className="flex flex-col gap-2">
      <h3 className="font-medium">Finish-to-start chain</h3>
      <ul aria-label="Dependency tree" className="flex flex-col gap-1 text-sm">
        {model.predecessors.length === 0 ? (
          <li className="text-muted-foreground">No predecessors.</li>
        ) : (
          model.predecessors.map((dep) => (
            <li key={dep.id} className="flex items-center justify-between gap-2">
              <span className="flex flex-wrap items-center gap-2">
                <span>{dep.name}</span>
                {dep.status ? <StatusBadge kind="task" value={dep.status} /> : null}
              </span>
              {predecessorAction?.(dep.id)}
            </li>
          ))
        )}
        <li className="pl-4">
          <span className="flex flex-wrap items-center gap-2 font-medium">
            <span>{model.current.name}</span>
            <StatusBadge kind="task" value={model.current.status} />
            <span className="text-xs font-normal text-muted-foreground">This task</span>
          </span>
          <ul className="mt-1 flex flex-col gap-1 pl-4" aria-label="Waiting on this task">
            {model.dependents.length === 0 ? (
              <li className="text-muted-foreground">No dependents.</li>
            ) : (
              model.dependents.map((dep) => (
                <li key={dep.id} className="flex flex-wrap items-center gap-2">
                  <span>{dep.name}</span>
                  {dep.status ? <StatusBadge kind="task" value={dep.status} /> : null}
                  <span className="text-xs text-muted-foreground">Waiting on this task</span>
                </li>
              ))
            )}
          </ul>
        </li>
      </ul>
    </div>
  )
}
