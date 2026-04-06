import { apiRequest } from '../client'
import type { ErrorLogDto } from '../../types/dto/errorLogs'

export const errorLogService = {
  getAll(limit = 100) {
    return apiRequest<ErrorLogDto[]>('/error-logs', {
      query: { limit },
    })
  },
}
