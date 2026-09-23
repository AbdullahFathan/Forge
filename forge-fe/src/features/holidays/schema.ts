import { z } from "zod"

import type { HolidayPatch, HolidayWrite } from "@/features/holidays/api/holidays"

const dateRe = /^\d{4}-\d{2}-\d{2}$/

export const holidaySchema = z.object({
  date: z.string().regex(dateRe, "Use YYYY-MM-DD"),
  name: z.string().trim().min(1, "Name is required").max(128, "Name is too long"),
})

export type HolidayValues = z.infer<typeof holidaySchema>

export function toHolidayCreate(values: HolidayValues): HolidayWrite {
  return { date: values.date, name: values.name.trim() }
}

export function toHolidayPatch(values: HolidayValues): HolidayPatch {
  return { date: values.date, name: values.name.trim() }
}
