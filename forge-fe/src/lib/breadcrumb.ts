import { create } from "zustand"

type BreadcrumbState = {
  detailLabel: string | null
  setDetailLabel: (label: string | null) => void
}

export const usePageCrumb = create<BreadcrumbState>((set) => ({
  detailLabel: null,
  setDetailLabel: (detailLabel) => set({ detailLabel }),
}))
