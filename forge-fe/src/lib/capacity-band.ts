import type { CapacityBand } from "@/types/resource"
import { OVER_ALLOCATED_WARNING } from "@/types/resource"

export function bandFromUtilization(percent: number): CapacityBand {
  if (percent < 80) return "GREEN"
  if (percent <= 100) return "YELLOW"
  return "RED"
}

export function sumUtilization(percents: number[]) {
  return percents.reduce((total, value) => total + value, 0)
}

export function isOverAllocated(overAllocated: boolean, warnings: readonly string[] = []) {
  return overAllocated || warnings.includes(OVER_ALLOCATED_WARNING)
}

export function bandClassName(band: string) {
  if (band === "YELLOW") return "bg-cap-full text-cap-full-foreground"
  if (band === "RED") return "bg-cap-over text-cap-over-foreground"
  return "bg-cap-available text-cap-available-foreground"
}

export function bandLabel(band: string) {
  if (band === "YELLOW") return "Full"
  if (band === "RED") return "Over"
  return "Available"
}
