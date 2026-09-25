import { getApiClient } from "@/services/api/client"
import { endpoints } from "@/services/api/endpoints"
import { compactParams } from "@/lib/search-params"
import type { Page } from "@/types/api"
import type {
  AllocationPatch,
  AllocationPublic,
  AllocationWrite,
  PersonPublic,
} from "@/types/resource"

export type PeopleListParams = {
  page: number
  pageSize: number
  q?: string
}

export async function listPeople(params: PeopleListParams) {
  const { data } = await getApiClient().get<Page<PersonPublic>>(endpoints.resourcesPeople, {
    params: compactParams(params),
  })
  return data
}

export type AllocationListParams = {
  page: number
  pageSize: number
  userId?: string
  projectId?: string
  from?: string
  to?: string
}

export async function listAllocations(params: AllocationListParams) {
  const { data } = await getApiClient().get<Page<AllocationPublic>>(endpoints.resourcesAllocations, {
    params: compactParams(params),
  })
  return data
}

export async function createAllocation(input: AllocationWrite) {
  const { data } = await getApiClient().post<AllocationPublic>(endpoints.resourcesAllocations, input)
  return data
}

export async function patchAllocation(id: string, input: AllocationPatch) {
  const { data } = await getApiClient().patch<AllocationPublic>(endpoints.resourceAllocation(id), input)
  return data
}

export async function deleteAllocation(id: string) {
  const { data } = await getApiClient().delete<{ status: string }>(endpoints.resourceAllocation(id))
  return data
}
