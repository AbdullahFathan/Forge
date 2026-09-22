import { CapacityLegend } from "@/components/common/capacity-cell"
import { FilterBar } from "@/components/common/filter-bar"
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
import type { Department } from "@/types/common"

const ALL = "all"

export function CapacityFilterBar({
  from,
  to,
  granularity,
  departmentId,
  skill,
  departments,
  canPickDepartment,
  onFrom,
  onTo,
  onGranularity,
  onDepartmentId,
  onSkill,
}: {
  from: string
  to: string
  granularity: "week" | "month"
  departmentId: string
  skill: string
  departments: Department[]
  canPickDepartment: boolean
  onFrom: (value: string) => void
  onTo: (value: string) => void
  onGranularity: (value: "week" | "month") => void
  onDepartmentId: (value: string) => void
  onSkill: (value: string) => void
}) {
  return (
    <div className="flex flex-col gap-2">
      <FilterBar>
        <div className="flex flex-col gap-1">
          <Label htmlFor="cap-from">From</Label>
          <Input id="cap-from" type="date" value={from} onChange={(event) => onFrom(event.target.value)} />
        </div>
        <div className="flex flex-col gap-1">
          <Label htmlFor="cap-to">To</Label>
          <Input id="cap-to" type="date" value={to} onChange={(event) => onTo(event.target.value)} />
        </div>
        <div className="flex flex-col gap-1">
          <Label>Granularity</Label>
          <Select value={granularity} onValueChange={(value) => onGranularity(value === "month" ? "month" : "week")}>
            <SelectTrigger className="w-36">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectGroup>
                <SelectItem value="week">Week</SelectItem>
                <SelectItem value="month">Month</SelectItem>
              </SelectGroup>
            </SelectContent>
          </Select>
        </div>
        {canPickDepartment ? (
          <div className="flex flex-col gap-1">
            <Label>Department</Label>
            <Select value={departmentId} onValueChange={(value) => onDepartmentId(value ?? ALL)}>
              <SelectTrigger className="w-48">
                <SelectValue placeholder="All departments" />
              </SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  <SelectItem value={ALL}>All departments</SelectItem>
                  {departments.map((department) => (
                    <SelectItem key={department.id} value={department.id}>
                      {department.name}
                    </SelectItem>
                  ))}
                </SelectGroup>
              </SelectContent>
            </Select>
          </div>
        ) : null}
        <div className="flex flex-col gap-1">
          <Label htmlFor="cap-skill">Skill</Label>
          <Input
            id="cap-skill"
            value={skill}
            placeholder="Optional"
            onChange={(event) => onSkill(event.target.value)}
          />
        </div>
      </FilterBar>
      <CapacityLegend />
    </div>
  )
}
