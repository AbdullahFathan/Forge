import { zodResolver } from "@hookform/resolvers/zod"
import { useForm } from "react-hook-form"

import { loginSchema, type LoginValues } from "@/features/auth/schema"
import { APP_NAME } from "@/lib/constants"
import { useLogin } from "@/features/auth/hooks/use-login"
import { ApiError } from "@/types/api"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { Field, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { Spinner } from "@/components/ui/spinner"

function loginMessage(error: unknown) {
  if (!(error instanceof ApiError)) return "Could not sign in. Try again."
  if (error.code === "INVALID_CREDENTIALS") return "Email or password is incorrect."
  if (error.code === "USER_INACTIVE") return "This account is inactive."
  if (error.code === "RATE_LIMITED") return "Too many attempts. Wait a minute and try again."
  return error.message
}

export function LoginForm() {
  const login = useLogin()
  const form = useForm<LoginValues>({
    resolver: zodResolver(loginSchema),
    defaultValues: { email: "", password: "" },
  })

  return (
    <Card className="w-full max-w-lg shadow-sm">
      <CardHeader>
        <CardTitle className="text-2xl tracking-tight">{APP_NAME}</CardTitle>
        <CardDescription>Sign in with the account an administrator created for you.</CardDescription>
      </CardHeader>
      <CardContent>
        <form
          className="flex flex-col gap-4"
          onSubmit={form.handleSubmit((values) => login.mutate(values))}
          noValidate
        >
          {login.error ? (
            <Alert variant="destructive">
              <AlertTitle>Sign in failed</AlertTitle>
              <AlertDescription>{loginMessage(login.error)}</AlertDescription>
            </Alert>
          ) : null}
          <FieldGroup>
            <Field data-invalid={Boolean(form.formState.errors.email) || undefined}>
              <FieldLabel htmlFor="email">Email</FieldLabel>
              <Input
                id="email"
                type="email"
                autoComplete="username"
                aria-invalid={Boolean(form.formState.errors.email)}
                {...form.register("email")}
              />
              <FieldError errors={[form.formState.errors.email]} />
            </Field>
            <Field data-invalid={Boolean(form.formState.errors.password) || undefined}>
              <FieldLabel htmlFor="password">Password</FieldLabel>
              <Input
                id="password"
                type="password"
                autoComplete="current-password"
                aria-invalid={Boolean(form.formState.errors.password)}
                {...form.register("password")}
              />
              <FieldError errors={[form.formState.errors.password]} />
            </Field>
          </FieldGroup>
          <Button type="submit" disabled={login.isPending}>
            {login.isPending ? <Spinner data-icon="inline-start" /> : null}
            Sign in
          </Button>
        </form>
      </CardContent>
    </Card>
  )
}
