import { getApiClient } from "@/services/api/client"
import { endpoints } from "@/services/api/endpoints"
import { taskFilterParams } from "@/lib/search-params"
import type { Page } from "@/types/api"
import type { TaskComment, TaskPublic } from "@/types/task"

export type TaskListParams = {
  page: number
  pageSize: number
  status?: string
  assignee?: string
  priority?: string
  dueFrom?: string
  dueTo?: string
  label?: string
  include?: string
}

export type TaskWrite = {
  parentTaskId?: string | null
  name: string
  description?: string
  status?: string
  priority?: string
  estimatedHours?: number | null
  startDate?: string | null
  dueDate?: string | null
  labels?: string[]
  position?: number
  assigneeIds?: string[]
}

export type TaskPatch = Partial<TaskWrite> & {
  clearParent?: boolean
  clearEstimatedHours?: boolean
  clearStartDate?: boolean
  clearDueDate?: boolean
}

export async function listProjectTasks(projectId: string, params: TaskListParams) {
  const { data } = await getApiClient().get<Page<TaskPublic>>(endpoints.projectTasks(projectId), {
    params: taskFilterParams(params),
  })
  return data
}

export async function createTask(projectId: string, input: TaskWrite) {
  const { data } = await getApiClient().post<TaskPublic>(endpoints.projectTasks(projectId), input)
  return data
}

export async function getTask(id: string, includeSubtasks = false) {
  const { data } = await getApiClient().get<TaskPublic>(endpoints.task(id), {
    params: includeSubtasks ? { include: "subtasks" } : undefined,
  })
  return data
}

export async function patchTask(id: string, input: TaskPatch) {
  const { data } = await getApiClient().patch<TaskPublic>(endpoints.task(id), input)
  return data
}

export async function deleteTask(id: string) {
  const { data } = await getApiClient().delete<{ status: string }>(endpoints.task(id))
  return data
}

export async function addTaskDependency(taskId: string, dependsOnTaskId: string) {
  const { data } = await getApiClient().post<TaskPublic>(endpoints.taskDependencies(taskId), {
    dependsOnTaskId,
  })
  return data
}

export async function removeTaskDependency(taskId: string, depId: string) {
  const { data } = await getApiClient().delete<{ status: string }>(
    endpoints.taskDependency(taskId, depId),
  )
  return data
}

export async function listTaskComments(taskId: string) {
  const { data } = await getApiClient().get<{ items: TaskComment[] }>(endpoints.taskComments(taskId))
  return data
}

export async function addTaskComment(taskId: string, body: string) {
  const { data } = await getApiClient().post<TaskComment>(endpoints.taskComments(taskId), { body })
  return data
}
