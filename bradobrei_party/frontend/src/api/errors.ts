import { ApiError } from './client'

export function getReadableErrorMessage(error: unknown, fallback: string) {
  if (error instanceof ApiError) {
    switch (error.status) {
      case 401:
        return 'Сессия истекла или токен недействителен. Войдите в систему снова.'
      case 403:
        return 'Недостаточно прав для выполнения этого действия.'
      case 404:
        return error.message || 'Запрошенные данные не найдены.'
      case 500:
        return 'На сервере произошла ошибка. Повторите попытку позже.'
      default:
        return error.message || fallback
    }
  }

  if (error instanceof Error) {
    return error.message
  }

  return fallback
}
