import { useEffect, useMemo, useState } from 'react'
import { useOutletContext } from 'react-router-dom'
import { DataTable, type TableColumn } from '../components/DataTable'
import { bookingService } from '../api/services/bookingService'
import { salonService } from '../api/services/salonService'
import { serviceService } from '../api/services/serviceService'
import type { UserDto } from '../types/dto/auth'
import type {
  BookingDto,
  CreateBookingRequestDto,
  EmployeeProfileSummaryDto,
  SalonDto,
  ServiceDto,
} from '../types/dto/entities'
import { formatBookingStatus, formatCurrency, formatDateTime } from '../shared/formatters'

type BookingScope = 'all' | 'my' | 'master'

interface AppShellContext {
  currentUser: UserDto | null
}

const initialForm: CreateBookingRequestDto = {
  start_time: '',
  salon_id: 0,
  master_id: undefined,
  service_ids: [],
  notes: '',
}

/** Значение `datetime-local` (YYYY-MM-DDTHH:mm) → RFC3339 UTC, как ждёт `time.Parse(RFC3339)` на backend. */
function datetimeLocalToRFC3339(value: string): string {
  if (!value) {
    return ''
  }
  const d = new Date(value)
  if (Number.isNaN(d.getTime())) {
    return ''
  }
  return d.toISOString().replace(/\.\d{3}Z$/, 'Z')
}

function getDefaultScope(role?: string): BookingScope {
  if (role === 'CLIENT') {
    return 'my'
  }

  if (role === 'BASIC_MASTER' || role === 'ADVANCED_MASTER') {
    return 'master'
  }

  return 'all'
}

const bookingColumns = (
  onConfirm: (booking: BookingDto) => void,
  onCancel: (booking: BookingDto) => void,
): Array<TableColumn<BookingDto>> => [
  { key: 'id', header: 'ID', render: (row) => row.id },
  { key: 'time', header: 'Визит', render: (row) => formatDateTime(row.start_time) },
  { key: 'salon', header: 'Салон', render: (row) => row.salon?.name || `#${row.salon_id}` },
  { key: 'client', header: 'Клиент', render: (row) => row.client?.full_name || `#${row.client_id}` },
  {
    key: 'master',
    header: 'Мастер',
    render: (row) => row.master?.full_name || (row.master_id ? `#${row.master_id}` : 'Не назначен'),
  },
  {
    key: 'services',
    header: 'Услуги',
    render: (row) => row.items?.map((item) => item.service?.name || `#${item.service_id}`).join(', ') || '—',
  },
  { key: 'price', header: 'Сумма', render: (row) => formatCurrency(row.total_price) },
  {
    key: 'status',
    header: 'Статус',
    render: (row) => <span className="status-pill">{formatBookingStatus(row.status)}</span>,
  },
  {
    key: 'actions',
    header: 'Действия',
    render: (row) => (
      <div className="table-actions">
        {(row.status === 'PENDING' || row.status === 'CONFIRMED') ? (
          <button type="button" className="ghost-button button-small" onClick={() => onConfirm(row)}>
            Подтвердить
          </button>
        ) : null}
        {row.status !== 'CANCELLED' && row.status !== 'COMPLETED' ? (
          <button type="button" className="danger-button button-small" onClick={() => onCancel(row)}>
            Отменить
          </button>
        ) : null}
      </div>
    ),
  },
]

export function BookingsPage() {
  const { currentUser } = useOutletContext<AppShellContext>()
  const [bookings, setBookings] = useState<BookingDto[]>([])
  const [salons, setSalons] = useState<SalonDto[]>([])
  const [services, setServices] = useState<ServiceDto[]>([])
  const [masters, setMasters] = useState<EmployeeProfileSummaryDto[]>([])
  const [scope, setScope] = useState<BookingScope>(getDefaultScope(currentUser?.role))
  const [form, setForm] = useState(initialForm)
  const [loading, setLoading] = useState(true)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')
  const [message, setMessage] = useState('')

  const scopeOptions = useMemo(() => {
    const options: Array<{ value: BookingScope; label: string }> = [{ value: 'my', label: 'Мои записи' }]

    if (currentUser?.role === 'BASIC_MASTER' || currentUser?.role === 'ADVANCED_MASTER' || currentUser?.role === 'ADMINISTRATOR') {
      options.push({ value: 'master', label: 'Записи мастера' })
    }

    if (currentUser?.role === 'ADMINISTRATOR' || currentUser?.role === 'ACCOUNTANT' || currentUser?.role === 'NETWORK_MANAGER') {
      options.push({ value: 'all', label: 'Все записи' })
    }

    return options
  }, [currentUser?.role])

  useEffect(() => {
    setScope((current) => {
      const nextDefault = getDefaultScope(currentUser?.role)
      if (scopeOptions.some((option) => option.value === current)) {
        return current
      }

      if (scopeOptions.some((option) => option.value === nextDefault)) {
        return nextDefault
      }

      return scopeOptions[0]?.value || 'my'
    })
  }, [currentUser?.role, scopeOptions])

  async function loadReferenceData() {
    try {
      const [salonsResponse, servicesResponse] = await Promise.all([
        salonService.getAll(),
        serviceService.getAll(),
      ])
      setSalons(salonsResponse)
      setServices(servicesResponse)
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : 'Не удалось загрузить справочники для бронирований.')
    }
  }

  async function loadBookings(selectedScope: BookingScope) {
    setLoading(true)
    setError('')
    try {
      let response: BookingDto[]
      if (selectedScope === 'all') {
        response = await bookingService.getAll()
      } else if (selectedScope === 'master') {
        response = await bookingService.getMaster()
      } else {
        response = await bookingService.getMy()
      }
      setBookings(response)
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : 'Не удалось загрузить бронирования.')
      setBookings([])
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadReferenceData()
  }, [])

  useEffect(() => {
    loadBookings(scope)
  }, [scope])

  useEffect(() => {
    if (!form.salon_id) {
      setMasters([])
      return
    }

    salonService
      .getMasters(form.salon_id)
      .then((response) => setMasters(response))
      .catch(() => setMasters([]))
  }, [form.salon_id])

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setSubmitting(true)
    setError('')
    setMessage('')

    if (form.service_ids.length === 0) {
      setError('Выберите хотя бы одну услугу — без этого бронирование не создаётся.')
      setSubmitting(false)
      return
    }

    try {
      const startRFC3339 = datetimeLocalToRFC3339(form.start_time)
      if (!startRFC3339) {
        setError('Укажите корректную дату и время визита.')
        setSubmitting(false)
        return
      }

      const payload: CreateBookingRequestDto = {
        ...form,
        start_time: startRFC3339,
        salon_id: Number(form.salon_id),
        master_id: form.master_id ? Number(form.master_id) : undefined,
        service_ids: form.service_ids.map(Number),
      }

      const created = await bookingService.create(payload)
      setMessage(`Бронирование #${created.id} создано.`)
      setForm(initialForm)
      setMasters([])
      await loadBookings(scope)
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : 'Не удалось создать бронирование.')
    } finally {
      setSubmitting(false)
    }
  }

  async function handleConfirm(booking: BookingDto) {
    try {
      await bookingService.confirm(booking.id)
      setMessage(`Бронирование #${booking.id} подтверждено.`)
      await loadBookings(scope)
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : 'Не удалось подтвердить бронирование.')
    }
  }

  async function handleCancel(booking: BookingDto) {
    try {
      await bookingService.cancel(booking.id)
      setMessage(`Бронирование #${booking.id} отменено.`)
      await loadBookings(scope)
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : 'Не удалось отменить бронирование.')
    }
  }

  return (
    <section className="page-section">
      <div className="page-header">
        <p className="eyebrow">Записи и загрузка</p>
        <h2>Бронирования</h2>
        <p className="section-description">
          Страница для создания новых визитов и операционной работы с уже созданными записями. Для бронирований на бэкенде нет полного CRUD, поэтому здесь используются реальные действия confirm/cancel.
        </p>
      </div>

      {message ? <div className="alert alert-success">{message}</div> : null}
      {error ? <div className="alert alert-error">{error}</div> : null}

      <form className="card-form card-form-grid" onSubmit={handleSubmit}>
        <label className="field">
          <span>Дата и время визита</span>
          <input
            type="datetime-local"
            value={form.start_time}
            onChange={(event) => setForm((current) => ({ ...current, start_time: event.target.value }))}
            required
          />
          <small className="field-hint">Время вашего браузера; на сервер уходит в формате RFC3339 (UTC).</small>
        </label>
        <label className="field">
          <span>Салон</span>
          <select
            value={form.salon_id || ''}
            onChange={(event) => setForm((current) => ({ ...current, salon_id: Number(event.target.value), master_id: undefined }))}
            required
          >
            <option value="">Выберите салон</option>
            {salons.map((salon) => (
              <option key={salon.id} value={salon.id}>
                {salon.name}
              </option>
            ))}
          </select>
        </label>
        <label className="field">
          <span>Мастер</span>
          <select
            value={form.master_id || ''}
            onChange={(event) =>
              setForm((current) => ({
                ...current,
                master_id: event.target.value ? Number(event.target.value) : undefined,
              }))
            }
          >
            <option value="">Назначить позже</option>
            {masters.map((master) => (
              <option key={master.user_id} value={master.user_id}>
                {master.user?.full_name || `Профиль ${master.id}`}
              </option>
            ))}
          </select>
        </label>
        <label className="field field-wide">
          <span>Комментарий</span>
          <textarea rows={3} value={form.notes} onChange={(event) => setForm((current) => ({ ...current, notes: event.target.value }))} placeholder="Например: важна работа с бородой и усами." />
        </label>
        <div className="field field-wide">
          <span>Услуги (обязательно — минимум одна)</span>
          {services.length === 0 ? (
            <p className="section-description field-hint">
              В справочнике нет услуг. Создайте услуги на странице «Услуги», затем обновите эту страницу.
            </p>
          ) : (
            <div className="checkbox-grid">
              {services.map((service) => {
                const isChecked = form.service_ids.includes(service.id)
                return (
                  <label key={service.id} className="checkbox-card">
                    <input
                      type="checkbox"
                      checked={isChecked}
                      onChange={(event) =>
                        setForm((current) => ({
                          ...current,
                          service_ids: event.target.checked
                            ? [...current.service_ids, service.id]
                            : current.service_ids.filter((serviceId) => serviceId !== service.id),
                        }))
                      }
                    />
                    <span>
                      <strong>{service.name}</strong>
                      <small>{formatCurrency(service.price)} • {service.duration_minutes} мин.</small>
                    </span>
                  </label>
                )
              })}
            </div>
          )}
        </div>
        <button
          type="submit"
          className="primary-button field-wide"
          disabled={submitting || services.length === 0}
        >
          {submitting ? 'Создаём бронирование...' : 'Создать бронирование'}
        </button>
      </form>

      <div className="filter-card">
        <label className="field">
          <span>Область просмотра</span>
          <select value={scope} onChange={(event) => setScope(event.target.value as BookingScope)}>
            {scopeOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </label>
      </div>

      <DataTable
        caption={loading ? 'Загружаем бронирования...' : 'Журнал бронирований'}
        columns={bookingColumns(handleConfirm, handleCancel)}
        rows={bookings}
        emptyText="Бронирования для выбранного режима пока отсутствуют."
      />
    </section>
  )
}
