import { useCallback, useState } from "react"

export function useLocalStorage<T>(key: string, initial: T) {
  const [value, setValue] = useState<T>(() => {
    const raw = localStorage.getItem(key)
    if (raw == null) return initial
    try {
      return JSON.parse(raw) as T
    } catch {
      return initial
    }
  })

  const update = useCallback(
    (next: T | ((current: T) => T)) => {
      setValue((current) => {
        const resolved = typeof next === "function" ? (next as (current: T) => T)(current) : next
        localStorage.setItem(key, JSON.stringify(resolved))
        return resolved
      })
    },
    [key],
  )

  return [value, update] as const
}
