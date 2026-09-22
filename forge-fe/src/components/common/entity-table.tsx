import { Fragment, useState } from "react"

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { Button } from "@/components/ui/button"
import { Empty, EmptyDescription, EmptyHeader, EmptyTitle } from "@/components/ui/empty"
import { Skeleton } from "@/components/ui/skeleton"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"

export type EntityColumn<T> = {
  key: string
  header: string
  cell: (row: T) => React.ReactNode
  className?: string
}

export function EntityTable<T>({
  columns,
  rows,
  rowKey,
  isLoading,
  error,
  emptyTitle,
  emptyDescription,
  emptyAction,
  page,
  pageSize,
  totalItems,
  onPageChange,
  renderExpanded,
}: {
  columns: EntityColumn<T>[]
  rows: T[]
  rowKey: (row: T) => string
  isLoading?: boolean
  error?: string | null
  emptyTitle: string
  emptyDescription?: string
  emptyAction?: React.ReactNode
  page: number
  pageSize: number
  totalItems: number
  onPageChange: (page: number) => void
  renderExpanded?: (row: T) => React.ReactNode
}) {
  const [expandedId, setExpandedId] = useState<string | null>(null)
  const pageCount = Math.max(1, Math.ceil(totalItems / pageSize))
  const colSpan = columns.length + (renderExpanded ? 1 : 0)

  if (error) {
    return (
      <Alert variant="destructive">
        <AlertTitle>Could not load data</AlertTitle>
        <AlertDescription>{error}</AlertDescription>
      </Alert>
    )
  }

  if (!isLoading && rows.length === 0) {
    return (
      <Empty>
        <EmptyHeader>
          <EmptyTitle>{emptyTitle}</EmptyTitle>
          {emptyDescription ? <EmptyDescription>{emptyDescription}</EmptyDescription> : null}
        </EmptyHeader>
        {emptyAction}
      </Empty>
    )
  }

  return (
    <div className="flex flex-col gap-3">
      <Table>
        <TableHeader>
          <TableRow>
            {renderExpanded ? <TableHead className="w-20 px-3 py-2">Detail</TableHead> : null}
            {columns.map((column) => (
              <TableHead key={column.key} className="px-3 py-2">
                {column.header}
              </TableHead>
            ))}
          </TableRow>
        </TableHeader>
        <TableBody>
          {isLoading
            ? Array.from({ length: 5 }, (_, index) => (
                <TableRow key={index}>
                  {Array.from({ length: colSpan }, (_, cell) => (
                    <TableCell key={cell} className="px-3 py-2">
                      <Skeleton className="h-4 w-full" />
                    </TableCell>
                  ))}
                </TableRow>
              ))
            : rows.map((row) => {
                const id = rowKey(row)
                const open = expandedId === id
                return (
                  <Fragment key={id}>
                    <TableRow>
                      {renderExpanded ? (
                        <TableCell className="px-3 py-2">
                          <Button
                            type="button"
                            variant="ghost"
                            size="sm"
                            onClick={() => setExpandedId(open ? null : id)}
                          >
                            {open ? "Hide" : "Show"}
                          </Button>
                        </TableCell>
                      ) : null}
                      {columns.map((column) => (
                        <TableCell key={column.key} className={column.className ?? "px-3 py-2"}>
                          {column.cell(row)}
                        </TableCell>
                      ))}
                    </TableRow>
                    {open && renderExpanded ? (
                      <TableRow>
                        <TableCell colSpan={colSpan} className="px-3 py-2">
                          {renderExpanded(row)}
                        </TableCell>
                      </TableRow>
                    ) : null}
                  </Fragment>
                )
              })}
        </TableBody>
      </Table>
      <div className="flex items-center justify-between gap-2">
        <p className="text-xs text-muted-foreground tabular-nums">
          {totalItems} total
        </p>
        <div className="flex items-center gap-2">
          <Button
            type="button"
            variant="outline"
            size="sm"
            disabled={page <= 1 || isLoading}
            onClick={() => onPageChange(page - 1)}
          >
            Previous
          </Button>
          <span className="text-xs tabular-nums text-muted-foreground">
            {page} / {pageCount}
          </span>
          <Button
            type="button"
            variant="outline"
            size="sm"
            disabled={page >= pageCount || isLoading}
            onClick={() => onPageChange(page + 1)}
          >
            Next
          </Button>
        </div>
      </div>
    </div>
  )
}
