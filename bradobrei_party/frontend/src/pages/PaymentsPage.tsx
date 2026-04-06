import { useEffect, useMemo, useState } from 'react'
import { useOutletContext } from 'react-router-dom'
import { getReadableErrorMessage } from '../api/errors'
import { bookingService } from '../api/services/bookingService'
import { paymentService } from '../api/services/paymentService'
import { DataTable, type TableColumn } from '../components/DataTable'
import { formatCurrency, formatDateTime, formatPaymentStatus } from '../shared/formatters'
import { paymentStatusOptions } from '../shared/options'
import type { AppShellOutletContext } from '../types/appShell'
import type { BookingDto, CreatePaymentRequestDto, PaymentDto } from '../types/dto/entities'

const initialForm: CreatePaymentRequestDto = {
  booking_id: 0,
  amount: 0,
  status: 'PENDING',
  external_transaction_id: '',
}

const paymentColumns: Array<TableColumn<PaymentDto>> = [
  { key: 'id', header: 'ID', render: (row) => row.id },
  { key: 'booking', header: 'Бронирование', render: (row) => `#${row.booking_id}` },
  { key: 'amount', header: 'Сумма', render: (row) => formatCurrency(row.amount) },
  {
    key: 'status',
    header: 'Статус',
    render: (row) => <span className="status-pill">{formatPaymentStatus(row.status)}</span>,
  },
  { key: 'external', header: 'Транзакция', render: (row) => row.external_transaction_id || '—' },
  { key: 'created', header: 'Создан', render: (row) => formatDateTime(row.created_at) },
]

export function PaymentsPage() {
  const { currentUser } = useOutletContext<AppShellOutletContext>()
  const [payments, setPayments] = useState<PaymentDto[]>([])
  const [bookings, setBookings] = useState<BookingDto[]>([])
  const [form, setForm] = useState(initialForm)
  const [loading, setLoading] = useState(true)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')
  const [message, setMessage] = useState('')

  const canCreatePayments =
    currentUser?.role === 'ADMINISTRATOR' || currentUser?.role === 'ACCOUNTANT'

  const availableBookings = useMemo(
    () => bookings.filter((booking) => !booking.payment),
    [bookings],
  )

  async function loadPageData() {
    setLoading(true)
    try {
      const [paymentsResponse, bookingsResponse] = await Promise.all([
        paymentService.getAll(),
        bookingService.getAll(),
      ])
      setPayments(paymentsResponse)
      setBookings(bookingsResponse)
    } catch (requestError) {
      setError(getReadableErrorMessage(requestError, 'Не удалось загрузить платежи.'))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void loadPageData()
  }, [])

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (!canCreatePayments) {
      return
    }

    setSubmitting(true)
    setError('')
    setMessage('')

    try {
      const created = await paymentService.create({
        ...form,
        booking_id: Number(form.booking_id),
        amount: Number(form.amount),
      })
      setMessage(`Платёж #${created.id} создан.`)
      setForm(initialForm)
      await loadPageData()
    } catch (requestError) {
      setError(getReadableErrorMessage(requestError, 'Не удалось создать платёж.'))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <section className="page-section">
      <div className="page-header">
        <p className="eyebrow">Финансовый контур</p>
        <h2>Платежи</h2>
        <p className="section-description">
          Журнал оплаты по бронированиям. Создание платежей доступно администратору и бухгалтеру,
          а управляющий сетью работает в режиме просмотра.
        </p>
      </div>

      {message ? <div className="alert alert-success">{message}</div> : null}
      {error ? <div className="alert alert-error">{error}</div> : null}

      {canCreatePayments ? (
        <form className="card-form card-form-grid" onSubmit={handleSubmit}>
          <label className="field">
            <span>Бронирование</span>
            <select
              value={form.booking_id || ''}
              onChange={(event) => {
                const bookingId = Number(event.target.value)
                const booking = availableBookings.find((item) => item.id === bookingId)
                setForm((current) => ({
                  ...current,
                  booking_id: bookingId,
                  amount: booking?.total_price || current.amount,
                }))
              }}
              required
            >
              <option value="">Выберите бронирование</option>
              {availableBookings.map((booking) => (
                <option key={booking.id} value={booking.id}>
                  #{booking.id} • {booking.salon?.name || `Салон ${booking.salon_id}`} •{' '}
                  {formatCurrency(booking.total_price)}
                </option>
              ))}
            </select>
          </label>
          <label className="field">
            <span>Сумма</span>
            <input
              type="number"
              min="0"
              value={form.amount || ''}
              onChange={(event) =>
                setForm((current) => ({ ...current, amount: Number(event.target.value) }))
              }
              required
            />
          </label>
          <label className="field">
            <span>Статус</span>
            <select
              value={form.status}
              onChange={(event) =>
                setForm((current) => ({
                  ...current,
                  status: event.target.value as CreatePaymentRequestDto['status'],
                }))
              }
            >
              {paymentStatusOptions.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
          </label>
          <label className="field">
            <span>Внешний ID транзакции</span>
            <input
              value={form.external_transaction_id}
              onChange={(event) =>
                setForm((current) => ({ ...current, external_transaction_id: event.target.value }))
              }
              placeholder="txn_local_12345"
            />
          </label>
          <button type="submit" className="primary-button field-wide" disabled={submitting}>
            {submitting ? 'Создаём платёж...' : 'Создать платёж'}
          </button>
        </form>
      ) : (
        <div className="filter-card">
          <div className="report-note">
            Для вашей роли доступен просмотр платежей без создания новых транзакций.
          </div>
        </div>
      )}

      <DataTable
        caption={loading ? 'Загружаем платежи...' : 'Журнал платежей'}
        columns={paymentColumns}
        rows={payments}
        emptyText="Платежи пока не заведены."
      />
    </section>
  )
}
