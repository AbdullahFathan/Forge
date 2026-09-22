import { Spinner } from "@/components/ui/spinner"

export function Loader({ label = "Loading" }: { label?: string }) {
  return (
    <div className="flex min-h-40 flex-1 items-center justify-center gap-2 text-sm text-muted-foreground">
      <Spinner />
      <span>{label}</span>
    </div>
  )
}
