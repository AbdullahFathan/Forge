import { Avatar, AvatarFallback, AvatarGroup, AvatarGroupCount } from "@/components/ui/avatar"

function initials(name: string) {
  return name
    .split(" ")
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part.charAt(0).toUpperCase())
    .join("")
}

export function UserPile({
  people,
  max = 3,
}: {
  people: { id: string; name: string }[]
  max?: number
}) {
  if (people.length === 0) {
    return <span className="text-muted-foreground">Unassigned</span>
  }
  const shown = people.slice(0, max)
  const extra = people.length - shown.length
  return (
    <AvatarGroup className="justify-start">
      {shown.map((person) => (
        <Avatar key={person.id} size="sm" title={person.name}>
          <AvatarFallback>{initials(person.name)}</AvatarFallback>
        </Avatar>
      ))}
      {extra > 0 ? <AvatarGroupCount>+{extra}</AvatarGroupCount> : null}
    </AvatarGroup>
  )
}
