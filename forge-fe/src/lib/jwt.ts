export type AccessClaims = {
  sub?: string
  role?: string
  perms?: string[]
}

export function decodeAccessToken(token: string): AccessClaims {
  const segment = token.split(".")[1]
  if (!segment) return {}
  try {
    const padded = segment.replace(/-/g, "+").replace(/_/g, "/")
    const json = atob(padded.padEnd(padded.length + ((4 - (padded.length % 4)) % 4), "="))
    return JSON.parse(json) as AccessClaims
  } catch {
    return {}
  }
}
