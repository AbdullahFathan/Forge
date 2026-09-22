export type FieldErrorDetail = {
  field: string
  tag: string
  message: string
}

export type ApiErrorBody = {
  code: string
  message: string
  details?: unknown
}

export type Envelope<T> = {
  success: boolean
  data?: T
  error?: ApiErrorBody
}

export type Page<T> = {
  items: T[]
  page: number
  pageSize: number
  totalItems: number
}

export class ApiError extends Error {
  code: string
  status: number
  details?: unknown

  constructor(status: number, body: ApiErrorBody) {
    super(body.message)
    this.name = "ApiError"
    this.code = body.code
    this.status = status
    this.details = body.details
  }
}
