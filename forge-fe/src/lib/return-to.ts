import { RETURN_TO_KEY } from "@/lib/constants"

export function rememberReturnTo(path: string) {
  sessionStorage.setItem(RETURN_TO_KEY, path)
}

export function consumeReturnTo() {
  const path = sessionStorage.getItem(RETURN_TO_KEY)
  sessionStorage.removeItem(RETURN_TO_KEY)
  if (!path || !path.startsWith("/") || path.startsWith("//")) return "/"
  return path
}
