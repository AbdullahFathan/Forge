import { Link } from "react-router"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { useEffect, useMemo, useState } from "react"
import { toast } from "sonner"

import { EntityTable } from "@/components/common/entity-table"
import { FilterBar } from "@/components/common/filter-bar"
import { PageHeader } from "@/components/common/page-header"
import { Button } from "@/components/ui/button"
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog"
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import {
  createHoliday,
  deleteHoliday,
  listHolidays,
  patchHoliday,
} from "@/features/holidays/api/holidays"
import { HolidayFormDialog } from "@/features/holidays/components/holiday-form-dialog"
import { toHolidayCreate, toHolidayPatch, type HolidayValues } from "@/features/holidays/schema"
import { can } from "@/lib/auth"
import { DEFAULT_PAGE_SIZE, PERMISSIONS } from "@/lib/constants"
import { useSession } from "@/lib/session"
import { ApiError } from "@/types/api"
import type { HolidayPublic } from "@/types/resource"
import { queryKeys } from "@/services/query/query-keys"

const ALL = "all"

function yearOf(date: string) {
  return date.slice(0, 4)
}

export function HolidaysPage() {
  const queryClient = useQueryClient()
  const permissions = useSession((state) => state.permissions)
  const canViewCapacity = can(permissions, PERMISSIONS.capacityView)
  const currentYear = String(new Date().getFullYear())
  const [page, setPage] = useState(1)
  const [year, setYear] = useState(ALL)
  const [open, setOpen] = useState(false)
  const [editing, setEditing] = useState<HolidayPublic | null>(null)
  const [removing, setRemoving] = useState<HolidayPublic | null>(null)

  const list = useQuery({
    queryKey: queryKeys.holidays,
    queryFn: listHolidays,
  })

  const years = useMemo(() => {
    const set = new Set<string>([currentYear])
    for (const row of list.data ?? []) {
      const y = yearOf(row.date)
      if (y) set.add(y)
    }
    return [...set].sort((a, b) => Number(b) - Number(a))
  }, [list.data, currentYear])

  const filtered = useMemo(() => {
    const rows = list.data ?? []
    if (year === ALL) return rows
    return rows.filter((row) => yearOf(row.date) === year)
  }, [list.data, year])

  useEffect(() => {
    const maxPage = Math.max(1, Math.ceil(filtered.length / DEFAULT_PAGE_SIZE))
    if (page > maxPage) setPage(maxPage)
  }, [filtered.length, page])

  const pageRows = filtered.slice((page - 1) * DEFAULT_PAGE_SIZE, page * DEFAULT_PAGE_SIZE)

  async function invalidateRelated() {
    await Promise.all([
      queryClient.invalidateQueries({ queryKey: ["resources"] }),
      queryClient.invalidateQueries({ queryKey: ["me"] }),
      queryClient.invalidateQueries({ queryKey: ["dashboards"] }),
    ])
  }

  const save = useMutation({
    mutationFn: (values: HolidayValues) =>
      editing
        ? patchHoliday(editing.id, toHolidayPatch(values))
        : createHoliday(toHolidayCreate(values)),
    onSuccess: async () => {
      await invalidateRelated()
      toast.success(editing ? "Holiday updated" : "Holiday created")
      setOpen(false)
      setEditing(null)
    },
    onError: (error) => {
      toast.error(error instanceof ApiError ? error.message : "Could not save holiday")
    },
  })

  const remove = useMutation({
    mutationFn: (id: string) => deleteHoliday(id),
    onSuccess: async () => {
      await invalidateRelated()
      toast.success("Holiday removed")
      setRemoving(null)
    },
    onError: (error) => {
      toast.error(error instanceof ApiError ? error.message : "Could not remove holiday")
    },
  })

  const error =
    list.error instanceof ApiError
      ? list.error.message
      : list.error
        ? "Could not load holidays"
        : null

  return (
    <div className="flex flex-col gap-4">
      <PageHeader
        title="Holidays"
        description="Dates on this list are the holidays used in the capacity formula. Weekends are already excluded. Personal leave is not modeled and is not subtracted. Allocated hours on a holiday still count, so utilization can rise."
        actions={
          <Button
            onClick={() => {
              setEditing(null)
              save.reset()
              setOpen(true)
            }}
          >
            New holiday
          </Button>
        }
      />
      {canViewCapacity ? (
        <p className="text-xs text-muted-foreground">
          <Link className="underline-offset-4 hover:underline" to="/resources">
            View capacity forecast
          </Link>
        </p>
      ) : null}
      <FilterBar>
        <Select
          value={year}
          onValueChange={(value) => {
            setYear(value ?? ALL)
            setPage(1)
          }}
        >
          <SelectTrigger className="w-36" aria-label="Filter by year">
            <SelectValue placeholder="All years" />
          </SelectTrigger>
          <SelectContent>
            <SelectGroup>
              <SelectItem value={ALL}>All years</SelectItem>
              {years.map((item) => (
                <SelectItem key={item} value={item}>
                  {item}
                </SelectItem>
              ))}
            </SelectGroup>
          </SelectContent>
        </Select>
      </FilterBar>
      <EntityTable
        columns={[
          { key: "date", header: "Date", className: "px-3 py-2 font-mono tabular-nums", cell: (row) => row.date },
          { key: "name", header: "Name", cell: (row) => row.name },
          {
            key: "actions",
            header: "",
            cell: (row) => (
              <div className="flex justify-end gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => {
                    setEditing(row)
                    save.reset()
                    setOpen(true)
                  }}
                >
                  Edit
                </Button>
                <Button variant="outline" size="sm" onClick={() => setRemoving(row)}>
                  Remove
                </Button>
              </div>
            ),
          },
        ]}
        rows={pageRows}
        rowKey={(row) => row.id}
        isLoading={list.isLoading}
        error={error}
        emptyTitle="No holidays yet"
        emptyDescription="Add public or company holidays so capacity uses fewer effective days. Personal leave is not part of this list."
        emptyAction={
          <Button
            onClick={() => {
              setEditing(null)
              setOpen(true)
            }}
          >
            New holiday
          </Button>
        }
        page={page}
        pageSize={DEFAULT_PAGE_SIZE}
        totalItems={filtered.length}
        onPageChange={setPage}
      />
      <HolidayFormDialog
        open={open}
        holiday={editing}
        pending={save.isPending}
        error={save.error}
        onOpenChange={setOpen}
        onSubmit={(values) => save.mutate(values)}
      />
      <AlertDialog open={Boolean(removing)} onOpenChange={(next) => !next && setRemoving(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Remove holiday</AlertDialogTitle>
            <AlertDialogDescription>
              {removing
                ? `Remove ${removing.name} on ${removing.date}? Capacity will treat that weekday as a working day again.`
                : ""}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              variant="destructive"
              disabled={remove.isPending}
              onClick={() => removing && remove.mutate(removing.id)}
            >
              Remove
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
