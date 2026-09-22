import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { useState } from "react"
import { toast } from "sonner"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"

import { StatusBadge } from "@/components/common/status-badge"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { Button } from "@/components/ui/button"
import { Field, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field"
import { Textarea } from "@/components/ui/textarea"
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet"
import {
  addTaskComment,
  addTaskDependency,
  listTaskComments,
  removeTaskDependency,
} from "@/features/tasks/api/tasks"
import { commentSchema, type CommentValues } from "@/features/tasks/schema"
import { flattenTasks } from "@/features/tasks/flatten"
import { ApiError } from "@/types/api"
import type { TaskPublic } from "@/types/task"
import { queryKeys } from "@/services/query/query-keys"

export function TaskSheet({
  task,
  allTasks,
  canComment,
  canManageDeps,
  onOpenChange,
}: {
  task: TaskPublic | null
  allTasks: TaskPublic[]
  canComment: boolean
  canManageDeps: boolean
  onOpenChange: (open: boolean) => void
}) {
  const queryClient = useQueryClient()
  const [dependsOnTaskId, setDependsOnTaskId] = useState("")
  const comments = useQuery({
    queryKey: queryKeys.taskComments(task?.id ?? ""),
    queryFn: () => listTaskComments(task!.id),
    enabled: Boolean(task),
  })
  const commentForm = useForm<CommentValues>({
    resolver: zodResolver(commentSchema),
    defaultValues: { body: "" },
  })

  const comment = useMutation({
    mutationFn: (values: CommentValues) => addTaskComment(task!.id, values.body),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: queryKeys.taskComments(task!.id) })
      commentForm.reset({ body: "" })
    },
    onError: (error) => {
      toast.error(error instanceof ApiError ? error.message : "Could not comment")
    },
  })

  const addDep = useMutation({
    mutationFn: (dependsOnTaskId: string) => addTaskDependency(task!.id, dependsOnTaskId),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["projects", task!.projectId, "tasks"] })
      setDependsOnTaskId("")
      toast.success("Dependency added")
    },
    onError: (error) => {
      toast.error(error instanceof ApiError ? error.message : "Could not add dependency")
    },
  })

  const removeDep = useMutation({
    mutationFn: (depId: string) => removeTaskDependency(task!.id, depId),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["projects", task!.projectId, "tasks"] })
      toast.success("Dependency removed")
    },
  })

  const candidates = flattenTasks(allTasks)
    .map(({ task: item }) => item)
    .filter((item) => item.id !== task?.id)

  return (
    <Sheet open={Boolean(task)} onOpenChange={onOpenChange}>
      <SheetContent className="overflow-y-auto sm:max-w-md">
        {task ? (
          <>
            <SheetHeader>
              <SheetTitle>{task.name}</SheetTitle>
              <SheetDescription>Finish-to-start predecessors must be Done before this task can start.</SheetDescription>
            </SheetHeader>
            <div className="flex flex-col gap-4 px-4 pb-6">
              <div className="flex flex-wrap items-center gap-2">
                <StatusBadge kind="task" value={task.status} />
              </div>
              {task.warnings.length > 0 ? (
                <Alert>
                  <AlertTitle>Dependency warnings</AlertTitle>
                  <AlertDescription>
                    <ul className="flex flex-col gap-1">
                      {task.warnings.map((warning) => (
                        <li key={warning}>{warning}</li>
                      ))}
                    </ul>
                  </AlertDescription>
                </Alert>
              ) : null}
              <div className="flex flex-col gap-2">
                <h3 className="font-medium">Depends on</h3>
                {task.dependsOn.length === 0 ? (
                  <p className="text-sm text-muted-foreground">No predecessors.</p>
                ) : (
                  <ul className="flex flex-col gap-1">
                    {task.dependsOn.map((dep) => (
                      <li key={dep.id} className="flex items-center justify-between gap-2 text-sm">
                        <span>
                          {dep.name ?? dep.taskId.slice(0, 8)}{" "}
                          {dep.status ? <StatusBadge kind="task" value={dep.status} /> : null}
                        </span>
                        {canManageDeps ? (
                          <Button variant="ghost" size="sm" onClick={() => removeDep.mutate(dep.id)}>
                            Remove
                          </Button>
                        ) : null}
                      </li>
                    ))}
                  </ul>
                )}
                {task.dependents.length > 0 ? (
                  <p className="text-sm text-muted-foreground">
                    Blocks {task.dependents.length} later task{task.dependents.length === 1 ? "" : "s"}.
                  </p>
                ) : null}
                {canManageDeps ? (
                  <div className="flex gap-2">
                    <Select value={dependsOnTaskId} onValueChange={(value) => setDependsOnTaskId(value ?? "")}>
                      <SelectTrigger className="flex-1" aria-label="Predecessor task">
                        <SelectValue placeholder="Add predecessor" />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectGroup>
                          {candidates.map((item) => (
                            <SelectItem key={item.id} value={item.id}>
                              {item.name}
                            </SelectItem>
                          ))}
                        </SelectGroup>
                      </SelectContent>
                    </Select>
                    <Button
                      type="button"
                      size="sm"
                      disabled={!dependsOnTaskId || addDep.isPending}
                      onClick={() => addDep.mutate(dependsOnTaskId)}
                    >
                      Add
                    </Button>
                  </div>
                ) : null}
              </div>
              <div className="flex flex-col gap-2">
                <h3 className="font-medium">Comments</h3>
                {(comments.data?.items ?? []).length === 0 ? (
                  <p className="text-sm text-muted-foreground">No comments yet.</p>
                ) : (
                  <ul className="flex flex-col gap-2">
                    {(comments.data?.items ?? []).map((item) => (
                      <li key={item.id} className="rounded-lg border p-2 text-sm">
                        <p className="text-xs text-muted-foreground">
                          {item.userName ?? item.userId} · {new Date(item.createdAt).toLocaleString()}
                        </p>
                        <p>{item.body}</p>
                      </li>
                    ))}
                  </ul>
                )}
                {canComment ? (
                  <form className="flex flex-col gap-2" onSubmit={commentForm.handleSubmit((values) => comment.mutate(values))}>
                    <FieldGroup>
                      <Field data-invalid={Boolean(commentForm.formState.errors.body) || undefined}>
                        <FieldLabel htmlFor="comment-body">Progress comment</FieldLabel>
                        <Textarea id="comment-body" rows={3} {...commentForm.register("body")} />
                        <FieldError errors={[commentForm.formState.errors.body]} />
                      </Field>
                    </FieldGroup>
                    <Button type="submit" size="sm" disabled={comment.isPending}>
                      Add comment
                    </Button>
                  </form>
                ) : null}
              </div>
            </div>
          </>
        ) : null}
      </SheetContent>
    </Sheet>
  )
}
