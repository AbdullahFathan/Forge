import { zodResolver } from "@hookform/resolvers/zod"
import { useEffect } from "react"
import { useForm } from "react-hook-form"

import { applyApiErrors } from "@/lib/validators"
import { holidaySchema, type HolidayValues } from "@/features/holidays/schema"
import { ApiError } from "@/types/api"
import type { HolidayPublic } from "@/types/resource"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Field, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { Spinner } from "@/components/ui/spinner"

export function HolidayFormDialog({
  open,
  holiday,
  pending,
  error,
  onOpenChange,
  onSubmit,
}: {
  open: boolean
  holiday: HolidayPublic | null
  pending: boolean
  error: unknown
  onOpenChange: (open: boolean) => void
  onSubmit: (values: HolidayValues) => void
}) {
  const form = useForm<HolidayValues>({
    resolver: zodResolver(holidaySchema),
    defaultValues: { date: "", name: "" },
  })

  useEffect(() => {
    if (!open) return
    form.reset({ date: holiday?.date ?? "", name: holiday?.name ?? "" })
  }, [open, holiday, form])

  useEffect(() => {
    applyApiErrors(error, form.setError)
  }, [error, form])

  const message = error instanceof ApiError ? error.message : null

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>{holiday ? "Edit holiday" : "New holiday"}</DialogTitle>
          <DialogDescription>
            Holidays drop weekdays from effective capacity. Allocated hours on that date still count.
          </DialogDescription>
        </DialogHeader>
        <form className="flex flex-col gap-4" onSubmit={form.handleSubmit(onSubmit)}>
          <FieldGroup>
            <Field data-invalid={Boolean(form.formState.errors.date) || undefined}>
              <FieldLabel htmlFor="holiday-date">Date</FieldLabel>
              <Input
                id="holiday-date"
                type="date"
                aria-invalid={Boolean(form.formState.errors.date)}
                {...form.register("date")}
              />
              <FieldError errors={[form.formState.errors.date]} />
            </Field>
            <Field data-invalid={Boolean(form.formState.errors.name) || undefined}>
              <FieldLabel htmlFor="holiday-name">Name</FieldLabel>
              <Input
                id="holiday-name"
                aria-invalid={Boolean(form.formState.errors.name)}
                {...form.register("name")}
              />
              <FieldError errors={[form.formState.errors.name]} />
            </Field>
          </FieldGroup>
          {message ? <p className="text-sm text-destructive">{message}</p> : null}
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={pending}>
              {pending ? <Spinner data-icon="inline-start" /> : null}
              Save
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
