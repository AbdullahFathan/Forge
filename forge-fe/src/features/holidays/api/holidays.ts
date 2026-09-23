import { getApiClient } from "@/services/api/client"
import { endpoints } from "@/services/api/endpoints"
import type { HolidayPublic } from "@/types/resource"

export type HolidayWrite = {
  date: string
  name: string
}

export type HolidayPatch = {
  date?: string
  name?: string
}

export async function listHolidays() {
  const { data } = await getApiClient().get<HolidayPublic[]>(endpoints.resourcesHolidays)
  return data
}

export async function createHoliday(input: HolidayWrite) {
  const { data } = await getApiClient().post<HolidayPublic>(endpoints.resourcesHolidays, input)
  return data
}

export async function patchHoliday(id: string, input: HolidayPatch) {
  const { data } = await getApiClient().patch<HolidayPublic>(endpoints.resourceHoliday(id), input)
  return data
}

export async function deleteHoliday(id: string) {
  const { data } = await getApiClient().delete<{ status: string }>(endpoints.resourceHoliday(id))
  return data
}
