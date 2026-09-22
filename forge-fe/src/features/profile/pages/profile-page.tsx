import { zodResolver } from "@hookform/resolvers/zod"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import { useEffect } from "react"
import { Controller, useForm } from "react-hook-form"
import { toast } from "sonner"

import { PageHeader } from "@/components/common/page-header"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Field, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { Spinner } from "@/components/ui/spinner"
import { Switch } from "@/components/ui/switch"
import { patchMe } from "@/features/profile/api/profile"
import { useCurrentUser } from "@/features/profile/hooks/use-current-user"
import { profileSchema, type ProfileValues } from "@/features/profile/schema"
import { formatRole } from "@/lib/auth"
import { applyApiErrors } from "@/lib/validators"
import { useSession } from "@/lib/session"
import { ApiError } from "@/types/api"
import { queryKeys } from "@/services/query/query-keys"

export function ProfilePage() {
  const queryClient = useQueryClient()
  const user = useSession((state) => state.user)
  const role = useSession((state) => state.role)
  const setUser = useSession((state) => state.setUser)
  const me = useCurrentUser()

  const form = useForm<ProfileValues>({
    resolver: zodResolver(profileSchema),
    defaultValues: {
      name: user?.name ?? "",
      currentPassword: "",
      password: "",
      emailNotificationsEnabled: user?.emailNotificationsEnabled ?? true,
    },
  })

  useEffect(() => {
    if (!user) return
    form.reset({
      name: user.name,
      currentPassword: "",
      password: "",
      emailNotificationsEnabled: user.emailNotificationsEnabled,
    })
  }, [user, form])

  const save = useMutation({
    mutationFn: (values: ProfileValues) =>
      patchMe({
        name: values.name,
        emailNotificationsEnabled: values.emailNotificationsEnabled,
        ...(values.password
          ? { password: values.password, currentPassword: values.currentPassword }
          : {}),
      }),
    onSuccess: async (next) => {
      setUser(next)
      await queryClient.invalidateQueries({ queryKey: queryKeys.me })
      toast.success("Profile updated")
      form.reset({
        name: next.name,
        currentPassword: "",
        password: "",
        emailNotificationsEnabled: next.emailNotificationsEnabled,
      })
    },
    onError: (error) => applyApiErrors(error, form.setError),
  })

  const message = save.error instanceof ApiError ? save.error.message : null

  return (
    <div className="flex flex-col gap-4">
      <PageHeader title="Profile" description="Your name, password, and email notification preference." />
      <Card className="max-w-lg shadow-sm">
        <CardHeader>
          <CardTitle className="text-sm font-medium">
            {user?.email} · {formatRole(role)}
            {user?.departmentName ? ` · ${user.departmentName}` : ""}
          </CardTitle>
        </CardHeader>
        <CardContent>
          <form className="flex flex-col gap-4" onSubmit={form.handleSubmit((values) => save.mutate(values))}>
            <FieldGroup>
              <Field data-invalid={Boolean(form.formState.errors.name) || undefined}>
                <FieldLabel htmlFor="profile-name">Name</FieldLabel>
                <Input id="profile-name" aria-invalid={Boolean(form.formState.errors.name)} {...form.register("name")} />
                <FieldError errors={[form.formState.errors.name]} />
              </Field>
              <Field data-invalid={Boolean(form.formState.errors.currentPassword) || undefined}>
                <FieldLabel htmlFor="profile-current">Current password</FieldLabel>
                <Input
                  id="profile-current"
                  type="password"
                  autoComplete="current-password"
                  aria-invalid={Boolean(form.formState.errors.currentPassword)}
                  {...form.register("currentPassword")}
                />
                <FieldError errors={[form.formState.errors.currentPassword]} />
              </Field>
              <Field data-invalid={Boolean(form.formState.errors.password) || undefined}>
                <FieldLabel htmlFor="profile-password">New password</FieldLabel>
                <Input
                  id="profile-password"
                  type="password"
                  autoComplete="new-password"
                  aria-invalid={Boolean(form.formState.errors.password)}
                  {...form.register("password")}
                />
                <FieldError errors={[form.formState.errors.password]} />
              </Field>
              <Field orientation="horizontal">
                <Controller
                  control={form.control}
                  name="emailNotificationsEnabled"
                  render={({ field }) => (
                    <Switch
                      id="profile-email"
                      checked={field.value}
                      onCheckedChange={field.onChange}
                    />
                  )}
                />
                <FieldLabel htmlFor="profile-email">Email notifications</FieldLabel>
              </Field>
            </FieldGroup>
            {message ? <p className="text-sm text-destructive">{message}</p> : null}
            <Button type="submit" disabled={save.isPending || me.isLoading}>
              {save.isPending ? <Spinner data-icon="inline-start" /> : null}
              Save profile
            </Button>
          </form>
        </CardContent>
      </Card>
    </div>
  )
}
