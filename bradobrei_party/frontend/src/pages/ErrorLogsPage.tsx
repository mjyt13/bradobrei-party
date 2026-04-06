import { useEffect, useState } from 'react'
import { getReadableErrorMessage } from '../api/errors'
import { errorLogService } from '../api/services/errorLogService'
import { DataTable, type TableColumn } from '../components/DataTable'
import { formatDateTime, formatRole } from '../shared/formatters'
import type { ErrorLogDto } from '../types/dto/errorLogs'

const columns: Array<TableColumn<ErrorLogDto>> = [
  { key: 'created_at', header: 'Время', render: (row) => formatDateTime(row.created_at) },
  { key: 'status_code', header: 'HTTP', render: (row) => row.status_code },
  { key: 'method', header: 'Метод', render: (row) => row.method },
  {
    key: 'path',
    header: 'Маршрут',
    render: (row) => (row.query ? `${row.path}?${row.query}` : row.path),
  },
  { key: 'error_code', header: 'Код ошибки', render: (row) => row.error_code || '—' },
  { key: 'message', header: 'Сообщение', render: (row) => row.message || '—' },
  {
    key: 'user_role',
    header: 'Роль',
    render: (row) => (row.user_role ? formatRole(row.user_role) : 'Гость'),
  },
  {
    key: 'user_id',
    header: 'Пользователь',
    render: (row) => (row.user_id ? `#${row.user_id}` : '—'),
  },
  { key: 'ip', header: 'IP', render: (row) => row.ip || '—' },
]

export function ErrorLogsPage() {
  const [logs, setLogs] = useState<ErrorLogDto[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [limit, setLimit] = useState(100)

  async function loadLogs(nextLimit = limit) {
    setLoading(true)
    setError('')

    try {
      setLogs(await errorLogService.getAll(nextLimit))
    } catch (requestError) {
      setError(getReadableErrorMessage(requestError, 'Не удалось загрузить журнал ошибок.'))
      setLogs([])
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void loadLogs()
  }, [])

  return (
    <section className="page-section">
      <div className="page-header">
        <p className="eyebrow">Исключительные ситуации</p>
        <h2>Журнал ошибок</h2>
        <p className="section-description">
          Здесь собираются ответы с кодами 400, 401, 403, 404 и 500. Журнал помогает быстро
          увидеть, что именно пошло не так, на каком маршруте и под какой ролью.
        </p>
      </div>

      {error ? <div className="alert alert-error">{error}</div> : null}

      <form
        className="filter-card"
        onSubmit={(event) => {
          event.preventDefault()
          void loadLogs(limit)
        }}
      >
        <label className="field">
          <span>Количество записей</span>
          <input
            type="number"
            min="1"
            max="500"
            value={limit}
            onChange={(event) => setLimit(Number(event.target.value))}
          />
        </label>
        <button type="submit" className="primary-button">
          {loading ? 'Обновляем...' : 'Показать журнал'}
        </button>
      </form>

      <DataTable
        caption={loading ? 'Загружаем журнал ошибок...' : 'Журнал исключительных ситуаций'}
        columns={columns}
        rows={logs}
        emptyText="Пока нет записей об ошибках."
      />
    </section>
  )
}
