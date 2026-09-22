import { useQuery } from "@tanstack/react-query"
import { useState } from "react"
import { toast } from "sonner"

import { EntityTable } from "@/components/common/entity-table"
import { FilterBar } from "@/components/common/filter-bar"
import { JsonDiff } from "@/components/common/json-diff"
import { PageHeader } from "@/components/common/page-header"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Spinner } from "@/components/ui/spinner"
import { listAuditLogs } from "@/features/audit/api/audit"
import { downloadBlob } from "@/lib/download"
import { compactParams } from "@/lib/search-params"
import { AUDIT_PAGE_SIZE } from "@/lib/constants"
import { queryKeys } from "@/services/query/query-keys"
import { endpoints } from "@/services/api/endpoints"
import { ApiError } from "@/types/api"
import { AUDIT_ENTITY_TYPES } from "@/types/audit"

const ALL = "all"

export function AuditPage() {
  const [page, setPage] = useState(1)
  const [userId, setUserId] = useState("")
  const [entityType, setEntityType] = useState(ALL)
  const [action, setAction] = useState(ALL)
  const [from, setFrom] = useState("")
  const [to, setTo] = useState("")
  const [downloading, setDownloading] = useState(false)

  const params = compactParams({
    page,
    pageSize: AUDIT_PAGE_SIZE,
    userId: userId.trim() || undefined,
    entityType: entityType === ALL ? undefined : entityType,
    action: action === ALL ? undefined : action,
    from: from || undefined,
    to: to || undefined,
  })

  const list = useQuery({
    queryKey: queryKeys.auditLogs(params),
    queryFn: () =>
      listAuditLogs({
        page,
        pageSize: AUDIT_PAGE_SIZE,
        userId: userId.trim() || undefined,
        entityType: entityType === ALL ? undefined : entityType,
        action: action === ALL ? undefined : action,
        from: from || undefined,
        to: to || undefined,
      }),
  })

  async function exportCsv() {
    setDownloading(true)
    try {
      await downloadBlob(
        endpoints.auditLogs,
        { ...params, format: "csv", page: 1, pageSize: 500 },
        "audit-logs.csv",
      )
    } catch (error) {
      toast.error(error instanceof ApiError ? error.message : "Could not export CSV")
    } finally {
      setDownloading(false)
    }
  }

  const error =
    list.error instanceof ApiError ? list.error.message : list.error ? "Could not load audit log" : null

  return (
    <div className="flex flex-col gap-4">
      <PageHeader
        title="Audit log"
        description="Immutable write history. Export is for compliance review only."
        actions={
          <Button type="button" variant="outline" disabled={downloading} onClick={() => void exportCsv()}>
            {downloading ? <Spinner data-icon="inline-start" /> : null}
            Export CSV
          </Button>
        }
      />
      <FilterBar>
        <div className="flex flex-col gap-1">
          <Label htmlFor="audit-from">From</Label>
          <Input id="audit-from" type="date" value={from} onChange={(event) => { setFrom(event.target.value); setPage(1) }} />
        </div>
        <div className="flex flex-col gap-1">
          <Label htmlFor="audit-to">To</Label>
          <Input id="audit-to" type="date" value={to} onChange={(event) => { setTo(event.target.value); setPage(1) }} />
        </div>
        <div className="flex flex-col gap-1">
          <Label htmlFor="audit-user">User ID</Label>
          <Input
            id="audit-user"
            value={userId}
            placeholder="UUID"
            onChange={(event) => {
              setUserId(event.target.value)
              setPage(1)
            }}
          />
        </div>
        <div className="flex flex-col gap-1">
          <Label>Entity</Label>
          <Select value={entityType} onValueChange={(value) => { setEntityType(value ?? ALL); setPage(1) }}>
            <SelectTrigger className="w-48">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectGroup>
                <SelectItem value={ALL}>All entities</SelectItem>
                {AUDIT_ENTITY_TYPES.map((type) => (
                  <SelectItem key={type} value={type}>
                    {type}
                  </SelectItem>
                ))}
              </SelectGroup>
            </SelectContent>
          </Select>
        </div>
        <div className="flex flex-col gap-1">
          <Label>Action</Label>
          <Select value={action} onValueChange={(value) => { setAction(value ?? ALL); setPage(1) }}>
            <SelectTrigger className="w-40">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectGroup>
                <SelectItem value={ALL}>All actions</SelectItem>
                <SelectItem value="CREATED">CREATED</SelectItem>
                <SelectItem value="UPDATED">UPDATED</SelectItem>
                <SelectItem value="DELETED">DELETED</SelectItem>
              </SelectGroup>
            </SelectContent>
          </Select>
        </div>
      </FilterBar>
      <EntityTable
        columns={[
          {
            key: "when",
            header: "When (UTC)",
            className: "px-3 py-2 font-mono tabular-nums",
            cell: (row) => (
              <time dateTime={row.createdAt} title={new Date(row.createdAt).toUTCString()}>
                {new Date(row.createdAt).toISOString()}
              </time>
            ),
          },
          { key: "user", header: "User", cell: (row) => row.userId },
          { key: "entity", header: "Entity", cell: (row) => `${row.entityType} · ${row.entityId.slice(0, 8)}` },
          { key: "action", header: "Action", cell: (row) => row.action },
          { key: "ip", header: "IP", cell: (row) => row.ipAddress || "—" },
        ]}
        rows={list.data?.items ?? []}
        rowKey={(row) => row.id}
        isLoading={list.isLoading}
        error={error}
        emptyTitle="No audit rows"
        emptyDescription="Write actions appear here after they succeed."
        page={page}
        pageSize={list.data?.pageSize ?? AUDIT_PAGE_SIZE}
        totalItems={list.data?.totalItems ?? 0}
        onPageChange={setPage}
        renderExpanded={(row) => <JsonDiff before={row.before} after={row.after} />}
      />
    </div>
  )
}
