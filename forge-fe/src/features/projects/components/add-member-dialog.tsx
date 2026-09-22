import { zodResolver } from "@hookform/resolvers/zod"
import { useEffect } from "react"
import { Controller, useForm } from "react-hook-form"
import { z } from "zod"

import { PROJECT_ROLES } from "@/lib/lifecycle"
import { applyApiErrors } from "@/lib/validators"
import { ApiError } from "@/types/api"
import type { PublicUser } from "@/types/user"
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
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Spinner } from "@/components/ui/spinner"

const schema = z.object({
  userId: z.string().min(1, "Select a user"),
  role: z.enum(PROJECT_ROLES),
})

export type AddMemberValues = z.infer<typeof schema>

export function AddMemberDialog({
  open,
  users,
  canPickUsers,
  pending,
  error,
  onOpenChange,
  onSubmit,
}: {
  open: boolean
  users: PublicUser[]
  canPickUsers: boolean
  pending: boolean
  error: unknown
  onOpenChange: (open: boolean) => void
  onSubmit: (values: AddMemberValues) => void
}) {
  const form = useForm<AddMemberValues>({
    resolver: zodResolver(schema),
    defaultValues: { userId: "", role: "MEMBER" },
  })

  useEffect(() => {
    applyApiErrors(error, form.setError)
  }, [error, form])

  const message = error instanceof ApiError ? error.message : null

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Add member</DialogTitle>
          <DialogDescription>Project roles are Lead, Member, or Viewer. Duplicate members are rejected.</DialogDescription>
        </DialogHeader>
        {canPickUsers ? (
          <form className="flex flex-col gap-4" onSubmit={form.handleSubmit(onSubmit)}>
            <FieldGroup>
              <Field data-invalid={Boolean(form.formState.errors.userId) || undefined}>
                <FieldLabel>User</FieldLabel>
                <Controller
                  control={form.control}
                  name="userId"
                  render={({ field }) => (
                    <Select value={field.value} onValueChange={(value) => field.onChange(value ?? "")}>
                      <SelectTrigger className="w-full" aria-invalid={Boolean(form.formState.errors.userId)}>
                        <SelectValue placeholder="Select a user" />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectGroup>
                          {users.map((user) => (
                            <SelectItem key={user.id} value={user.id}>
                              {user.name} ({user.email})
                            </SelectItem>
                          ))}
                        </SelectGroup>
                      </SelectContent>
                    </Select>
                  )}
                />
                <FieldError errors={[form.formState.errors.userId]} />
              </Field>
              <Field>
                <FieldLabel>Project role</FieldLabel>
                <Controller
                  control={form.control}
                  name="role"
                  render={({ field }) => (
                    <Select value={field.value} onValueChange={(value) => field.onChange(value ?? "MEMBER")}>
                      <SelectTrigger className="w-full">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectGroup>
                          {PROJECT_ROLES.map((role) => (
                            <SelectItem key={role} value={role}>
                              {role}
                            </SelectItem>
                          ))}
                        </SelectGroup>
                      </SelectContent>
                    </Select>
                  )}
                />
              </Field>
            </FieldGroup>
            {message ? <p className="text-sm text-destructive">{message}</p> : null}
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
                Cancel
              </Button>
              <Button type="submit" disabled={pending}>
                {pending ? <Spinner data-icon="inline-start" /> : null}
                Add
              </Button>
            </DialogFooter>
          </form>
        ) : (
          <p className="text-sm text-muted-foreground">
            Adding members needs the org user list (`user.manage`). Ask an administrator to add people, or grant that permission.
          </p>
        )}
      </DialogContent>
    </Dialog>
  )
}
