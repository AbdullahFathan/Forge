import type { ComponentProps } from "react"
import type { LucideIcon } from "lucide-react"
import {
  Archive,
  ChevronsDown,
  ChevronsUp,
  Circle,
  CircleCheck,
  CirclePlay,
  Equal,
  File,
  Flame,
  Inbox,
  Loader,
  OctagonAlert,
  Pause,
  ScanEye,
} from "lucide-react"

import type { Badge } from "@/components/ui/badge"

type BadgeVariant = ComponentProps<typeof Badge>["variant"]

export type StatusVisual = {
  label: string
  icon: LucideIcon
  variant: BadgeVariant
  className?: string
}

export const projectStatusMap: Record<string, StatusVisual> = {
  DRAFT: { label: "Draft", icon: File, variant: "secondary" },
  ACTIVE: { label: "Active", icon: CirclePlay, variant: "default" },
  ON_HOLD: { label: "On hold", icon: Pause, variant: "outline", className: "text-warning" },
  COMPLETED: { label: "Completed", icon: CircleCheck, variant: "outline", className: "text-success" },
  ARCHIVED: { label: "Archived", icon: Archive, variant: "secondary" },
}

export const taskStatusMap: Record<string, StatusVisual> = {
  BACKLOG: { label: "Backlog", icon: Inbox, variant: "secondary" },
  TODO: { label: "To do", icon: Circle, variant: "outline" },
  IN_PROGRESS: { label: "In progress", icon: Loader, variant: "default" },
  IN_REVIEW: { label: "In review", icon: ScanEye, variant: "outline" },
  DONE: { label: "Done", icon: CircleCheck, variant: "outline", className: "text-success" },
  BLOCKED: { label: "Blocked", icon: OctagonAlert, variant: "outline", className: "text-destructive" },
}

export const priorityMap: Record<string, StatusVisual> = {
  LOW: { label: "Low", icon: ChevronsDown, variant: "secondary" },
  MEDIUM: { label: "Medium", icon: Equal, variant: "outline" },
  HIGH: { label: "High", icon: ChevronsUp, variant: "outline", className: "text-warning" },
  CRITICAL: { label: "Critical", icon: Flame, variant: "destructive" },
}
