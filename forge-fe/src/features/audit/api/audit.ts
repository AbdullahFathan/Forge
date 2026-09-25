import { getApiClient } from "@/services/api/client"
import { endpoints } from "@/services/api/endpoints"
import { compactParams } from "@/lib/search-params"
import type { Page } from "@/types/api"
import type { AuditLogPublic } from "@/types/audit"

export type AuditListParams = {
  page: number
  pageSize: number
  userName?: string
  entityType?: string
  action?: string
  from?: string
  to?: string
}

export async function listAuditLogs(params: AuditListParams) {
  const { data } = await getApiClient().get<Page<AuditLogPublic>>(endpoints.auditLogs, {
    params: compactParams(params),
  })
  return data
}
