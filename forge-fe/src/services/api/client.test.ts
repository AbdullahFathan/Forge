import { AxiosError, type AxiosResponse } from "axios"
import { describe, expect, it, vi } from "vitest"

import { createApiClient, unwrapEnvelope } from "@/services/api/client"
import { ApiError } from "@/types/api"

describe("unwrapEnvelope", () => {
  it("returns data when success is true", () => {
    expect(unwrapEnvelope({ success: true, data: { id: "1" } })).toEqual({ id: "1" })
  })

  it("throws ApiError when success is false", () => {
    expect(() =>
      unwrapEnvelope({
        success: false,
        error: { code: "VALIDATION_ERROR", message: "invalid" },
      }),
    ).toThrow(ApiError)
  })
})

describe("401 retry", () => {
  it("refreshes once and retries the original request", async () => {
    const tokens = {
      current: "old-token",
      get() {
        return this.current
      },
      set(token: string | null) {
        this.current = token ?? ""
      },
    }
    const refresh = vi.fn(async () => "new-token")
    const client = createApiClient({
      baseURL: "http://api.test",
      tokens,
      refresh,
    })

    let calls = 0
    client.defaults.adapter = async (config) => {
      calls += 1
      const auth = config.headers.get("Authorization")
      if (calls === 1) {
        expect(auth).toBe("Bearer old-token")
        throw new AxiosError(
          "expired",
          "ERR_BAD_REQUEST",
          config,
          null,
          {
            data: { success: false, error: { code: "UNAUTHORIZED", message: "expired" } },
            status: 401,
            statusText: "Unauthorized",
            headers: {},
            config,
          },
        )
      }
      expect(auth).toBe("Bearer new-token")
      return {
        data: { success: true, data: { ok: true } },
        status: 200,
        statusText: "OK",
        headers: {},
        config,
      } as AxiosResponse
    }

    const response = await client.get("/secure")
    expect(response.data).toEqual({ ok: true })
    expect(refresh).toHaveBeenCalledTimes(1)
    expect(tokens.current).toBe("new-token")
  })
})
