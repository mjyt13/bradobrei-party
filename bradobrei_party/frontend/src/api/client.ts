import { apiConfig } from './config'
import { tokenStorage } from './services/tokenStorage'
import type { ErrorResponseDto } from '../types/dto/common'

type Primitive = string | number | boolean | null | undefined
type QueryParams = Record<string, Primitive>

interface RequestOptions extends Omit<RequestInit, 'body'> {
  auth?: boolean
  body?: unknown
  query?: QueryParams
}

export class ApiError extends Error {
  status: number
  payload?: ErrorResponseDto

  constructor(status: number, message: string, payload?: ErrorResponseDto) {
    super(message)
    this.status = status
    this.payload = payload
  }
}

function notifyUnauthorized() {
  window.dispatchEvent(new CustomEvent('auth:unauthorized'))
}

function getApiErrorMessage(status: number, payload: ErrorResponseDto | undefined, fallback: string) {
  if (payload?.message) {
    return payload.message
  }

  if (payload?.error) {
    return payload.error
  }

  switch (status) {
    case 401:
      return 'Сессия истекла или доступ запрещён. Выполните вход повторно.'
    case 403:
      return 'Недостаточно прав для выполнения этого действия.'
    case 404:
      return 'Запрошенные данные не найдены.'
    case 500:
      return 'На сервере произошла ошибка. Повторите попытку позже.'
    default:
      return fallback
  }
}

function buildUrl(path: string, query?: QueryParams) {
  const normalizedPath = path.startsWith('/') ? path : `/${path}`
  const url = new URL(`${apiConfig.apiBaseUrl}${normalizedPath}`, window.location.origin)

  if (query) {
    for (const [key, value] of Object.entries(query)) {
      if (value !== undefined && value !== null && value !== '') {
        url.searchParams.set(key, String(value))
      }
    }
  }

  return url.toString()
}

export function buildApiUrl(path: string, query?: QueryParams) {
  return buildUrl(path, query)
}

async function parseResponse(response: Response) {
  const contentType = response.headers.get('content-type') || ''
  if (contentType.includes('application/json')) {
    return response.json()
  }
  return response.text()
}

export async function apiRequest<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const { auth = true, body, headers, query, ...rest } = options
  const token = tokenStorage.get()
  const requestHeaders = new Headers(headers)

  if (!requestHeaders.has('Accept')) {
    requestHeaders.set('Accept', 'application/json')
  }

  if (body !== undefined) {
    requestHeaders.set('Content-Type', 'application/json')
  }

  if (auth && token) {
    requestHeaders.set('Authorization', `Bearer ${token}`)
  }

  const response = await fetch(buildUrl(path, query), {
    ...rest,
    headers: requestHeaders,
    body: body !== undefined ? JSON.stringify(body) : undefined,
  })

  const payload = await parseResponse(response)

  if (!response.ok) {
    if (response.status === 401) {
      tokenStorage.clear()
      notifyUnauthorized()
    }

    const errorPayload = typeof payload === 'object' && payload !== null ? (payload as ErrorResponseDto) : undefined
    const message = getApiErrorMessage(response.status, errorPayload, response.statusText)
    throw new ApiError(response.status, message, errorPayload)
  }

  return payload as T
}

export async function apiBlobRequest(path: string, options: RequestOptions = {}): Promise<Blob> {
  const { auth = true, body, headers, query, ...rest } = options
  const token = tokenStorage.get()
  const requestHeaders = new Headers(headers)

  if (body !== undefined) {
    requestHeaders.set('Content-Type', 'application/json')
  }

  if (auth && token) {
    requestHeaders.set('Authorization', `Bearer ${token}`)
  }

  const response = await fetch(buildUrl(path, query), {
    ...rest,
    headers: requestHeaders,
    body: body !== undefined ? JSON.stringify(body) : undefined,
  })

  if (!response.ok) {
    let payload: unknown
    const contentType = response.headers.get('content-type') || ''
    if (contentType.includes('application/json')) {
      payload = await response.json()
    } else {
      payload = await response.text()
    }

    if (response.status === 401) {
      tokenStorage.clear()
      notifyUnauthorized()
    }

    const errorPayload = typeof payload === 'object' && payload !== null ? (payload as ErrorResponseDto) : undefined
    const message = getApiErrorMessage(
      response.status,
      errorPayload,
      typeof payload === 'string' ? payload : response.statusText,
    )
    throw new ApiError(response.status, message, errorPayload)
  }

  return response.blob()
}
