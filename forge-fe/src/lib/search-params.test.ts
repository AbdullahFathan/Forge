import { describe, expect, it } from "vitest"

import { compactParams, taskFilterParams } from "@/lib/search-params"

describe("compactParams", () => {
  it("drops empty all sentinels", () => {
    expect(compactParams({ status: "all", page: 1, q: "" })).toEqual({ page: 1 })
  })
})

describe("taskFilterParams", () => {
  it("keeps dueFrom and dueTo", () => {
    expect(
      taskFilterParams({
        page: 1,
        pageSize: 20,
        dueFrom: "2026-01-01",
        dueTo: "2026-01-31",
        status: "all",
      }),
    ).toEqual({
      page: 1,
      pageSize: 20,
      dueFrom: "2026-01-01",
      dueTo: "2026-01-31",
    })
  })
})
