import { useEffect, useState } from "react"
import { useQuery } from "@tanstack/react-query"
import { ChevronDownIcon } from "lucide-react"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"
import { Spinner } from "@/components/ui/spinner"
import { DEFAULT_PAGE_SIZE } from "@/lib/constants"
import { cn } from "@/lib/utils"
import { ApiError } from "@/types/api"
import type { Page } from "@/types/api"

export function PaginatedSearchSelect<T>({
  value,
  displayLabel,
  placeholder = "Select",
  searchPlaceholder = "Search",
  disabled,
  invalid,
  className,
  allowClear,
  clearLabel = "All",
  pageSize = DEFAULT_PAGE_SIZE,
  queryKeyPrefix,
  fetchPage,
  getId,
  getLabel,
  getDescription,
  onChange,
}: {
  value: string
  displayLabel?: string
  placeholder?: string
  searchPlaceholder?: string
  disabled?: boolean
  invalid?: boolean
  className?: string
  allowClear?: boolean
  clearLabel?: string
  pageSize?: number
  queryKeyPrefix: readonly unknown[]
  fetchPage: (args: { q: string; page: number; pageSize: number }) => Promise<Page<T>>
  getId: (item: T) => string
  getLabel: (item: T) => string
  getDescription?: (item: T) => string | undefined
  onChange: (id: string, item: T | null) => void
}) {
  const [open, setOpen] = useState(false)
  const [q, setQ] = useState("")
  const [debounced, setDebounced] = useState("")
  const [page, setPage] = useState(1)
  const [selectedLabel, setSelectedLabel] = useState(displayLabel ?? "")

  useEffect(() => {
    if (displayLabel) {
      setSelectedLabel(displayLabel)
      return
    }
    if (!value) setSelectedLabel("")
  }, [displayLabel, value])

  useEffect(() => {
    const timer = window.setTimeout(() => setDebounced(q.trim()), 300)
    return () => window.clearTimeout(timer)
  }, [q])

  useEffect(() => {
    setPage(1)
  }, [debounced])

  useEffect(() => {
    if (!open) {
      setQ("")
      setDebounced("")
      setPage(1)
    }
  }, [open])

  const list = useQuery({
    queryKey: [...queryKeyPrefix, { q: debounced, page, pageSize }],
    queryFn: () => fetchPage({ q: debounced, page, pageSize }),
    enabled: open && !disabled,
  })

  const totalItems = list.data?.totalItems ?? 0
  const totalPages = Math.max(1, Math.ceil(totalItems / pageSize))
  const shown = value ? selectedLabel || placeholder : placeholder

  return (
    <Popover open={open} onOpenChange={setOpen} modal={false}>
      <PopoverTrigger
        disabled={disabled}
        render={
          <Button
            type="button"
            variant="outline"
            disabled={disabled}
            aria-invalid={invalid || undefined}
            className={cn(
              "w-full justify-between font-normal",
              !value && "text-muted-foreground",
              className,
            )}
          />
        }
      >
        <span className="truncate">{shown}</span>
        <ChevronDownIcon className="size-4 text-muted-foreground" />
      </PopoverTrigger>
      <PopoverContent className="p-2" align="start">
        <Input
          value={q}
          onChange={(event) => setQ(event.target.value)}
          placeholder={searchPlaceholder}
          autoFocus
        />
        <div className="mt-1 max-h-56 overflow-y-auto">
          {allowClear ? (
            <button
              type="button"
              className="flex w-full rounded-md px-2 py-1.5 text-left text-sm hover:bg-accent hover:text-accent-foreground"
              onMouseDown={(event) => event.preventDefault()}
              onClick={(event) => {
                event.preventDefault()
                event.stopPropagation()
                setSelectedLabel("")
                onChange("", null)
                setOpen(false)
              }}
            >
              {clearLabel}
            </button>
          ) : null}
          {list.isLoading ? (
            <div className="flex items-center justify-center gap-2 py-6 text-sm text-muted-foreground">
              <Spinner />
              Loading
            </div>
          ) : list.error ? (
            <p className="px-2 py-4 text-sm text-destructive">
              {list.error instanceof ApiError ? list.error.message : "Could not load options"}
            </p>
          ) : (list.data?.items ?? []).length === 0 ? (
            <p className="px-2 py-4 text-sm text-muted-foreground">No matches</p>
          ) : (
            (list.data?.items ?? []).map((item) => {
              const id = getId(item)
              const label = getLabel(item)
              const description = getDescription?.(item)
              return (
                <button
                  key={id}
                  type="button"
                  className={cn(
                    "flex w-full flex-col rounded-md px-2 py-1.5 text-left hover:bg-accent hover:text-accent-foreground",
                    id === value && "bg-accent",
                  )}
                  onMouseDown={(event) => event.preventDefault()}
                  onClick={(event) => {
                    event.preventDefault()
                    event.stopPropagation()
                    setSelectedLabel(label)
                    onChange(id, item)
                    setOpen(false)
                  }}
                >
                  <span className="text-sm">{label}</span>
                  {description ? (
                    <span className="text-xs text-muted-foreground">{description}</span>
                  ) : null}
                </button>
              )
            })
          )}
        </div>
        <div className="mt-1 flex items-center justify-between border-t pt-1 text-xs text-muted-foreground">
          <span>
            Page {page} of {totalPages}
          </span>
          <div className="flex gap-1">
            <Button
              type="button"
              size="xs"
              variant="ghost"
              disabled={page <= 1 || list.isFetching}
              onClick={() => setPage((current) => Math.max(1, current - 1))}
            >
              Prev
            </Button>
            <Button
              type="button"
              size="xs"
              variant="ghost"
              disabled={page >= totalPages || list.isFetching}
              onClick={() => setPage((current) => current + 1)}
            >
              Next
            </Button>
          </div>
        </div>
      </PopoverContent>
    </Popover>
  )
}
