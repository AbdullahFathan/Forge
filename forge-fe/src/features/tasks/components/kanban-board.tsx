import {
  DndContext,
  type DragEndEvent,
  PointerSensor,
  closestCorners,
  useDroppable,
  useSensor,
  useSensors,
} from "@dnd-kit/core"
import { CSS } from "@dnd-kit/utilities"
import { useSortable } from "@dnd-kit/sortable"
import { SortableContext, verticalListSortingStrategy } from "@dnd-kit/sortable"

import { PriorityBadge } from "@/components/common/priority-badge"
import { UserPile } from "@/components/common/user-pile"
import { taskStatusMap } from "@/components/common/status-map"
import { TASK_STATUSES, nextTaskStatuses } from "@/lib/lifecycle"
import type { TaskPublic } from "@/types/task"
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { cn } from "@/lib/utils"

function KanbanCard({
  task,
  reducedMotion,
  canChangeStatus,
  onOpen,
  onStatus,
}: {
  task: TaskPublic
  reducedMotion: boolean
  canChangeStatus: boolean
  onOpen: () => void
  onStatus: (status: string) => void
}) {
  const sortable = useSortable({ id: task.id, disabled: reducedMotion || !canChangeStatus })
  const style = {
    transform: CSS.Transform.toString(sortable.transform),
    transition: sortable.transition,
  }
  const options = nextTaskStatuses(task.status)

  return (
    <article
      ref={sortable.setNodeRef}
      style={style}
      {...sortable.attributes}
      {...sortable.listeners}
      className={cn(
        "flex flex-col gap-2 rounded-lg border bg-card p-3 text-sm",
        task.status === "BLOCKED" ? "border-destructive/50" : null,
      )}
    >
      <button type="button" className="text-left font-medium" onClick={onOpen}>
        {task.name}
      </button>
      <div className="flex items-center justify-between gap-2">
        <PriorityBadge value={task.priority} />
        {task.dueDate ? <span className="tabular-nums text-xs text-muted-foreground">{task.dueDate}</span> : null}
      </div>
      <UserPile people={task.assignees} />
      {canChangeStatus ? (
        <Select value={task.status} onValueChange={(value) => value && onStatus(value)}>
          <SelectTrigger size="sm" className="w-full" aria-label={`Status for ${task.name}`}>
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectGroup>
              {options.map((status) => (
                <SelectItem key={status} value={status}>
                  {taskStatusMap[status]?.label ?? status}
                </SelectItem>
              ))}
            </SelectGroup>
          </SelectContent>
        </Select>
      ) : null}
    </article>
  )
}

function Column({
  status,
  tasks,
  reducedMotion,
  canChangeStatus,
  onOpen,
  onStatus,
}: {
  status: string
  tasks: TaskPublic[]
  reducedMotion: boolean
  canChangeStatus: (task: TaskPublic) => boolean
  onOpen: (task: TaskPublic) => void
  onStatus: (task: TaskPublic, status: string) => void
}) {
  const { setNodeRef } = useDroppable({ id: status })
  const visual = taskStatusMap[status]
  return (
    <section className="flex min-w-64 flex-1 flex-col gap-2 rounded-lg bg-muted/40 p-2">
      <h3 className="px-1 text-sm font-medium">{visual?.label ?? status}</h3>
      <div ref={setNodeRef} className="flex min-h-24 flex-col gap-2">
        <SortableContext items={tasks.map((task) => task.id)} strategy={verticalListSortingStrategy}>
          {tasks.map((task) => (
            <KanbanCard
              key={task.id}
              task={task}
              reducedMotion={reducedMotion}
              canChangeStatus={canChangeStatus(task)}
              onOpen={() => onOpen(task)}
              onStatus={(value) => onStatus(task, value)}
            />
          ))}
        </SortableContext>
      </div>
    </section>
  )
}

export function KanbanBoard({
  tasks,
  reducedMotion,
  canChangeStatus,
  onOpen,
  onMove,
}: {
  tasks: TaskPublic[]
  reducedMotion: boolean
  canChangeStatus: (task: TaskPublic) => boolean
  onOpen: (task: TaskPublic) => void
  onMove: (taskId: string, status: string, position: number) => void
}) {
  const sensors = useSensors(useSensor(PointerSensor, { activationConstraint: { distance: 6 } }))
  const byStatus = Object.fromEntries(
    TASK_STATUSES.map((status) => [
      status,
      tasks.filter((task) => task.status === status).sort((a, b) => a.position - b.position),
    ]),
  ) as Record<string, TaskPublic[]>

  function handleDragEnd(event: DragEndEvent) {
    const { active, over } = event
    if (!over) return
    const task = tasks.find((item) => item.id === active.id)
    if (!task) return
    const overId = String(over.id)
    const overTask = tasks.find((item) => item.id === overId)
    const status = overTask?.status ?? (TASK_STATUSES.includes(overId as (typeof TASK_STATUSES)[number]) ? overId : task.status)
    const column = byStatus[status] ?? []
    const index = overTask ? column.findIndex((item) => item.id === overTask.id) : column.length
    onMove(task.id, status, Math.max(0, index))
  }

  const board = (
    <div className="flex gap-3 overflow-x-auto pb-2">
      {TASK_STATUSES.map((status) => (
        <Column
          key={status}
          status={status}
          tasks={byStatus[status] ?? []}
          reducedMotion={reducedMotion}
          canChangeStatus={canChangeStatus}
          onOpen={onOpen}
          onStatus={(task, value) => onMove(task.id, value, task.position)}
        />
      ))}
    </div>
  )

  if (reducedMotion) return board

  return (
    <DndContext sensors={sensors} collisionDetection={closestCorners} onDragEnd={handleDragEnd}>
      {board}
    </DndContext>
  )
}
