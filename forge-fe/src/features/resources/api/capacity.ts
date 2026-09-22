import { getApiClient } from "@/services/api/client"
import { endpoints } from "@/services/api/endpoints"
import { compactParams } from "@/lib/search-params"
import type {
  AllocationMatrix,
  AvailabilityItem,
  CapacityForecast,
  OverloadAlert,
  Workload,
} from "@/types/resource"

export type CapacityQuery = {
  from: string
  to: string
  granularity?: "week" | "month"
  departmentId?: string
  skill?: string
}

export async function getCapacity(params: CapacityQuery) {
  const { data } = await getApiClient().get<CapacityForecast>(endpoints.resourcesCapacity, {
    params: compactParams(params),
  })
  return data
}

export async function getMatrix(params: CapacityQuery) {
  const { data } = await getApiClient().get<AllocationMatrix>(endpoints.resourcesMatrix, {
    params: compactParams(params),
  })
  return data
}

export async function getAvailability(params: CapacityQuery) {
  const { data } = await getApiClient().get<AvailabilityItem[]>(endpoints.resourcesAvailability, {
    params: compactParams(params),
  })
  return data
}

export async function getOverloadAlerts() {
  const { data } = await getApiClient().get<OverloadAlert[]>(endpoints.resourcesOverloadAlerts)
  return data
}

export async function getWorkload(userId?: string) {
  const { data } = await getApiClient().get<Workload>(endpoints.meWorkload, {
    params: compactParams({ userId }),
  })
  return data
}
