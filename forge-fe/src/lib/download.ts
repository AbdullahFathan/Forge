import { getApiClient } from "@/services/api/client"
import { ApiError, type Envelope } from "@/types/api"

function filenameFromDisposition(header: string | undefined, fallback: string) {
  if (!header) return fallback
  const star = /filename\*=UTF-8''([^;]+)/i.exec(header)
  if (star?.[1]) return decodeURIComponent(star[1])
  const quoted = /filename="([^"]+)"/i.exec(header)
  if (quoted?.[1]) return quoted[1]
  const plain = /filename=([^;]+)/i.exec(header)
  if (plain?.[1]) return plain[1].trim()
  return fallback
}

function saveBlob(blob: Blob, filename: string) {
  const url = URL.createObjectURL(blob)
  const link = document.createElement("a")
  link.href = url
  link.download = filename
  link.click()
  URL.revokeObjectURL(url)
}

export async function downloadBlob(
  url: string,
  params: Record<string, unknown>,
  fallbackName: string,
) {
  const response = await getApiClient().get<Blob>(url, {
    params,
    responseType: "blob",
  })
  const blob = response.data
  const contentType = String(response.headers["content-type"] ?? "")
  if (contentType.includes("application/json")) {
    const parsed = JSON.parse(await blob.text()) as Envelope<unknown>
    throw new ApiError(response.status, parsed.error ?? { code: "INTERNAL_ERROR", message: "Download failed" })
  }
  const filename = filenameFromDisposition(
    String(response.headers["content-disposition"] ?? ""),
    fallbackName,
  )
  saveBlob(blob, filename)
}
