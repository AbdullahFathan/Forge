import type { FieldValues, Path, UseFormSetError } from "react-hook-form"

import { ApiError, type FieldErrorDetail } from "@/types/api"

const fieldAliases: Record<string, string> = {
  name: "name",
  email: "email",
  password: "password",
  roleid: "roleId",
  departmentid: "departmentId",
  capacityhoursperday: "capacityHoursPerDay",
  isactive: "isActive",
  currentpassword: "currentPassword",
  emailnotificationsenabled: "emailNotificationsEnabled",
  skills: "skills",
  startdate: "startDate",
  targetenddate: "targetEndDate",
  ownerid: "ownerId",
  tags: "tags",
  parenttaskid: "parentTaskId",
  estimatedhours: "estimatedHours",
  duedate: "dueDate",
  labels: "labels",
  assigneeids: "assigneeIds",
  dependsontaskid: "dependsOnTaskId",
  body: "body",
}

export function fieldKey(goField: string) {
  const compact = goField.replace(/[^a-zA-Z]/g, "").toLowerCase()
  return fieldAliases[compact] ?? goField.charAt(0).toLowerCase() + goField.slice(1)
}

export function isFieldErrorList(details: unknown): details is FieldErrorDetail[] {
  return (
    Array.isArray(details) &&
    details.every(
      (item) =>
        item &&
        typeof item === "object" &&
        "field" in item &&
        typeof item.field === "string",
    )
  )
}

export function applyApiErrors<T extends FieldValues>(
  error: unknown,
  setError: UseFormSetError<T>,
) {
  if (!(error instanceof ApiError) || !isFieldErrorList(error.details)) return
  for (const detail of error.details) {
    const key = fieldKey(detail.field) as Path<T>
    setError(key, { message: detail.message })
  }
}
