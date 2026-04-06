import type { UserRole } from './common'

export interface ErrorLogDto {
  id: number
  method: string
  path: string
  query?: string
  status_code: number
  error_code?: string
  message?: string
  ip?: string
  user_agent?: string
  user_id?: number
  user_role?: UserRole
  created_at: string
}
