import { useQuery } from "@tanstack/react-query"
import { useMemo, useState } from "react"

import { PageHeader } from "@/components/common/page-header"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { listDepartments } from "@/features/departments/api/departments"
import { AllocationsPanel } from "@/features/resources/components/allocations-panel"
import { AvailabilityPanel } from "@/features/resources/components/availability-panel"
import { CapacityFilterBar } from "@/features/resources/components/capacity-filter-bar"
import { ForecastPanel } from "@/features/resources/components/forecast-panel"
import { MatrixPanel } from "@/features/resources/components/matrix-panel"
import { listUsers } from "@/features/users/api/users"
import { can } from "@/lib/auth"
import { PERMISSIONS } from "@/lib/constants"
import { defaultCapacityRange } from "@/lib/date-range"
import { compactParams } from "@/lib/search-params"
import { useSession } from "@/lib/session"
import { queryKeys } from "@/services/query/query-keys"
import type { AllocationFormValues } from "@/features/resources/schema"
import type { AvailabilityItem } from "@/types/resource"
import type { CapacityQuery } from "@/features/resources/api/capacity"

const ALL = "all"

export function ResourcesPage() {
  const permissions = useSession((state) => state.permissions)
  const canAllocate = can(permissions, PERMISSIONS.resourceAllocate)
  const canView = can(permissions, PERMISSIONS.capacityView)
  const canPickUsers = can(permissions, PERMISSIONS.userManage)
  const canPickDepartment = can(permissions, PERMISSIONS.departmentManage)

  const range = defaultCapacityRange()
  const [tab, setTab] = useState(canAllocate ? "allocations" : "forecast")
  const [from, setFrom] = useState(range.from)
  const [to, setTo] = useState(range.to)
  const [granularity, setGranularity] = useState<"week" | "month">("week")
  const [departmentId, setDepartmentId] = useState(ALL)
  const [skill, setSkill] = useState("")
  const [formDefaults, setFormDefaults] = useState<Partial<AllocationFormValues> | null>(null)

  const users = useQuery({
    queryKey: queryKeys.users({ page: 1, pageSize: 100, picker: true }),
    queryFn: () => listUsers({ page: 1, pageSize: 100 }),
    enabled: canPickUsers,
  })

  const departments = useQuery({
    queryKey: queryKeys.departments({ page: 1, pageSize: 100, picker: true }),
    queryFn: () => listDepartments(1, 100),
    enabled: canPickDepartment && canView,
  })

  const capacityParams = useMemo<CapacityQuery>(
    () =>
      compactParams({
        from,
        to,
        granularity,
        departmentId: departmentId === ALL ? undefined : departmentId,
        skill: skill.trim() || undefined,
      }) as CapacityQuery,
    [from, to, granularity, departmentId, skill],
  )

  const onAllocate = (row: AvailabilityItem) => {
    setFormDefaults({ userId: row.userId })
    setTab("allocations")
  }

  return (
    <div className="flex flex-col gap-4">
      <PageHeader
        title="Resources"
        description="Allocations, capacity forecast, matrix, and availability. Over-allocation is saved as a warning, not blocked."
      />
      <Tabs
        value={tab}
        onValueChange={(value) => {
          if (typeof value === "string") setTab(value)
        }}
      >
        <TabsList variant="line">
          {canAllocate ? <TabsTrigger value="allocations">Allocations</TabsTrigger> : null}
          {canView ? <TabsTrigger value="forecast">Forecast</TabsTrigger> : null}
          {canView ? <TabsTrigger value="matrix">Matrix</TabsTrigger> : null}
          {canView ? <TabsTrigger value="availability">Availability</TabsTrigger> : null}
        </TabsList>
        {canAllocate ? (
          <TabsContent value="allocations">
            <AllocationsPanel
              users={users.data?.items ?? []}
              canPickUsers={canPickUsers}
              formDefaults={formDefaults}
              onFormDefaultsConsumed={() => setFormDefaults(null)}
            />
          </TabsContent>
        ) : null}
        {canView && tab !== "allocations" ? (
          <CapacityFilterBar
            from={from}
            to={to}
            granularity={granularity}
            departmentId={departmentId}
            skill={skill}
            departments={departments.data?.items ?? []}
            canPickDepartment={canPickDepartment}
            onFrom={setFrom}
            onTo={setTo}
            onGranularity={setGranularity}
            onDepartmentId={setDepartmentId}
            onSkill={setSkill}
          />
        ) : null}
        {canView ? (
          <>
            <TabsContent value="forecast">
              <ForecastPanel params={capacityParams} />
            </TabsContent>
            <TabsContent value="matrix">
              <MatrixPanel params={capacityParams} />
            </TabsContent>
            <TabsContent value="availability">
              <AvailabilityPanel
                params={capacityParams}
                canAllocate={canAllocate}
                onAllocate={onAllocate}
              />
            </TabsContent>
          </>
        ) : null}
      </Tabs>
    </div>
  )
}
