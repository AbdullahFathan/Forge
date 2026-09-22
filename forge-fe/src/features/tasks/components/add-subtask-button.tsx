import { canAddSubtask } from "@/lib/lifecycle"
import { Button } from "@/components/ui/button"

export function AddSubtaskButton({
  parentTaskId,
  onClick,
}: {
  parentTaskId: string | null | undefined
  onClick: () => void
}) {
  if (!canAddSubtask(parentTaskId)) return null
  return (
    <Button type="button" variant="ghost" size="sm" onClick={onClick}>
      Add subtask
    </Button>
  )
}
