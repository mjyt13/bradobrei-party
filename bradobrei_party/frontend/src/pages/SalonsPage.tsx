import { useEffect, useMemo, useState } from 'react'
import { useOutletContext } from 'react-router-dom'
import { ApiError } from '../api/client'
import { getReadableErrorMessage } from '../api/errors'
import { salonService } from '../api/services/salonService'
import { useConfirmDialog } from '../components/ConfirmDialog'
import { DataTable, type TableColumn } from '../components/DataTable'
import { SalonMap } from '../components/SalonMap'
import { ScheduleEditor } from '../components/ScheduleEditor'
import { salonStatusOptions } from '../shared/options'
import {
  formatCurrency,
  formatJsonPreview,
  formatRole,
  formatSalonStatus,
} from '../shared/formatters'
import type { AppShellOutletContext } from '../types/appShell'
import type {
  EmployeeProfileSummaryDto,
  SalonDto,
  UpsertSalonRequestDto,
} from '../types/dto/entities'

const initialForm: UpsertSalonRequestDto = {
  name: '',
  address: '',
  location: '',
  working_hours: '{"mon":"10:00-20:00","tue":"10:00-20:00"}',
  status: 'OPEN',
  max_staff: 8,
  base_hourly_rate: 1400,
}

const masterColumns: Array<TableColumn<EmployeeProfileSummaryDto>> = [
  {
    key: 'name',
    header: 'Сотрудник',
    render: (row) => row.user?.full_name || `Профиль #${row.id}`,
  },
  {
    key: 'role',
    header: 'Роль',
    render: (row) => formatRole(row.user?.role || '—'),
  },
  {
    key: 'spec',
    header: 'Специализация',
    render: (row) => row.specialization || '—',
  },
  {
    key: 'services',
    header: 'Услуги',
    render: (row) => row.services?.map((service) => service.name).join(', ') || '—',
  },
]

export function SalonsPage() {
  const { currentUser } = useOutletContext<AppShellOutletContext>()
  const { confirm, dialog } = useConfirmDialog()
  const [form, setForm] = useState(initialForm)
  const [salons, setSalons] = useState<SalonDto[]>([])
  const [masters, setMasters] = useState<EmployeeProfileSummaryDto[]>([])
  const [selectedSalon, setSelectedSalon] = useState<SalonDto | null>(null)
  const [editingId, setEditingId] = useState<number | null>(null)
  const [loading, setLoading] = useState(true)
  const [submitting, setSubmitting] = useState(false)
  const [geocodeSubmitting, setGeocodeSubmitting] = useState(false)
  const [error, setError] = useState('')
  const [message, setMessage] = useState('')

  const canManageSalons =
    currentUser?.role === 'ADMINISTRATOR' || currentUser?.role === 'NETWORK_MANAGER'
  const canDeleteSalons = currentUser?.role === 'ADMINISTRATOR'

  const columns = useMemo<Array<TableColumn<SalonDto>>>(
    () => [
      { key: 'name', header: 'Салон', render: (row) => row.name },
      { key: 'address', header: 'Адрес', render: (row) => row.address },
      {
        key: 'status',
        header: 'Статус',
        render: (row) => <span className="status-pill">{formatSalonStatus(row.status)}</span>,
      },
      { key: 'staff', header: 'Штат', render: (row) => `${row.max_staff} чел.` },
      {
        key: 'rate',
        header: 'Базовая ставка',
        render: (row) => formatCurrency(row.base_hourly_rate),
      },
      { key: 'hours', header: 'Часы', render: (row) => formatJsonPreview(row.working_hours) },
      {
        key: 'actions',
        header: 'Действия',
        render: (row) => (
          <div className="table-actions">
            {canManageSalons ? (
              <button
                type="button"
                className="ghost-button button-small"
                onClick={() => startEditing(row)}
              >
                Изменить
              </button>
            ) : null}
            <button
              type="button"
              className="ghost-button button-small"
              onClick={() => void handleShowMasters(row)}
            >
              Мастера
            </button>
            {canDeleteSalons ? (
              <button
                type="button"
                className="danger-button button-small"
                onClick={() => void handleDelete(row)}
              >
                Удалить
              </button>
            ) : null}
            {!canManageSalons && !canDeleteSalons ? <span>—</span> : null}
          </div>
        ),
      },
    ],
    [canDeleteSalons, canManageSalons],
  )

  useEffect(() => {
    void loadSalons()
  }, [])

  async function loadSalons() {
    setLoading(true)
    try {
      setSalons(await salonService.getAll())
    } catch (requestError) {
      setError(getReadableErrorMessage(requestError, 'Не удалось загрузить салоны.'))
    } finally {
      setLoading(false)
    }
  }

  function resetForm() {
    setEditingId(null)
    setForm(initialForm)
  }

  function startEditing(salon: SalonDto) {
    if (!canManageSalons) {
      setError('Недостаточно прав для редактирования салонов.')
      return
    }

    setEditingId(salon.id)
    setForm({
      name: salon.name,
      address: salon.address,
      location: salon.location || '',
      working_hours: salon.working_hours || '',
      status: salon.status,
      max_staff: salon.max_staff,
      base_hourly_rate: salon.base_hourly_rate,
    })
  }

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setSubmitting(true)
    setError('')
    setMessage('')

    if (!canManageSalons) {
      setError('Недостаточно прав для управления салонами.')
      setSubmitting(false)
      return
    }

    try {
      const payload = {
        ...form,
        working_hours: form.working_hours?.trim() || undefined,
        max_staff: Number(form.max_staff),
        base_hourly_rate: Number(form.base_hourly_rate),
      }

      if (editingId) {
        await salonService.update(editingId, payload)
        setMessage(`Салон #${editingId} обновлён.`)
      } else {
        const created = await salonService.create(payload)
        setMessage(`Салон "${created.name}" создан.`)
      }

      resetForm()
      await loadSalons()
    } catch (requestError) {
      setError(getReadableErrorMessage(requestError, 'Не удалось сохранить салон.'))
    } finally {
      setSubmitting(false)
    }
  }

  async function handleDelete(salon: SalonDto) {
    if (!canDeleteSalons) {
      setError('Удаление салонов доступно только администратору.')
      return
    }

    const shouldContinue = await confirm({
      title: 'Удаление салона',
      message: `Удалить салон "${salon.name}"?`,
      confirmLabel: 'Удалить',
      variant: 'danger',
    })

    if (!shouldContinue) {
      return
    }

    setError('')
    setMessage('')

    try {
      await salonService.remove(salon.id)
      setMessage(`Салон "${salon.name}" удалён.`)
      if (selectedSalon?.id === salon.id) {
        setSelectedSalon(null)
        setMasters([])
      }
      await loadSalons()
    } catch (requestError) {
      setError(getReadableErrorMessage(requestError, 'Не удалось удалить салон.'))
    }
  }

  async function handleGeocodeAddress() {
    setGeocodeSubmitting(true)
    setError('')
    setMessage('')

    try {
      const response = await salonService.geocodeAddress(form.address)
      setForm((current) => ({
        ...current,
        location: `${response.latitude}, ${response.longitude}`,
      }))
      setMessage(`Адрес подтверждён (${response.provider}): ${response.formatted_address}`)
    } catch (requestError) {
      if (requestError instanceof ApiError && requestError.status === 503) {
        setError(
          'На сервере не включён геокодер. Задайте GEOCODER_PROVIDER и ключи в backend .env или используйте ручные координаты.',
        )
      } else {
        setError(getReadableErrorMessage(requestError, 'Не удалось геокодировать адрес.'))
      }
    } finally {
      setGeocodeSubmitting(false)
    }
  }

  async function handleShowMasters(salon: SalonDto) {
    setSelectedSalon(salon)
    setError('')

    try {
      setMasters(await salonService.getMasters(salon.id))
    } catch (requestError) {
      setError(getReadableErrorMessage(requestError, 'Не удалось загрузить мастеров салона.'))
      setMasters([])
    }
  }

  return (
    <section className="page-section">
      {dialog}
      <div className="page-header">
        <p className="eyebrow">Операционный контур</p>
        <h2>Салоны</h2>
        <p className="section-description">
          Здесь можно вести справочник филиалов, обновлять параметры работы и быстро смотреть
          мастеров по выбранному салону.
        </p>
      </div>

      <SalonMap salons={salons} />

      {message ? <div className="alert alert-success">{message}</div> : null}
      {error ? <div className="alert alert-error">{error}</div> : null}

      {canManageSalons ? (
        <form className="card-form card-form-grid" onSubmit={handleSubmit}>
          <label className="field">
            <span>Название салона</span>
            <input
              value={form.name}
              onChange={(event) =>
                setForm((current) => ({ ...current, name: event.target.value }))
              }
              placeholder="Bradobrei Party Center"
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
                  status: event.target.value as UpsertSalonRequestDto['status'],
                }))
              }
            >
              {salonStatusOptions.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
          </label>
          <label className="field field-wide">
            <span>Адрес</span>
            <input
              value={form.address}
              onChange={(event) =>
                setForm((current) => ({ ...current, address: event.target.value }))
              }
              placeholder="Екатеринбург, ул. Малышева, 12"
              required
            />
          </label>
          <label className="field field-wide">
            <span>Координаты</span>
            <div className="field-inline">
              <input
                value={form.location || ''}
                onChange={(event) =>
                  setForm((current) => ({ ...current, location: event.target.value }))
                }
                placeholder="58.0141, 56.2230"
              />
              <button
                type="button"
                className="ghost-button"
                onClick={() => void handleGeocodeAddress()}
                disabled={geocodeSubmitting || !form.address.trim()}
              >
                {geocodeSubmitting ? 'Проверка...' : 'Проверить адрес на сервере'}
              </button>
            </div>
            <span className="field-hint">
              При включённом серверном геокодере координаты при сохранении будут уточняться по
              адресу.
            </span>
          </label>
          <label className="field">
            <span>Макс. персонал</span>
            <input
              type="number"
              min="1"
              value={form.max_staff}
              onChange={(event) =>
                setForm((current) => ({
                  ...current,
                  max_staff: Number(event.target.value),
                }))
              }
            />
          </label>
          <label className="field">
            <span>Базовая ставка</span>
            <input
              type="number"
              min="0"
              value={form.base_hourly_rate}
              onChange={(event) =>
                setForm((current) => ({
                  ...current,
                  base_hourly_rate: Number(event.target.value),
                }))
              }
            />
          </label>

          <ScheduleEditor
            label="Часы работы"
            value={form.working_hours}
            onChange={(nextValue) => setForm((current) => ({ ...current, working_hours: nextValue }))}
            hint="Оставьте день пустым, если салон в этот день не работает."
            placeholder="10:00-20:00"
          />

          <div className="button-row field-wide">
            <button type="submit" className="primary-button" disabled={submitting}>
              {submitting ? 'Сохраняем...' : editingId ? 'Обновить салон' : 'Создать салон'}
            </button>
            {editingId ? (
              <button type="button" className="ghost-button" onClick={resetForm}>
                Сбросить редактирование
              </button>
            ) : null}
          </div>
        </form>
      ) : (
        <div className="alert alert-info">
          Для вашей роли доступен просмотр салонов и мастеров, но не управление филиалами.
        </div>
      )}

      <DataTable
        caption={loading ? 'Загружаем салоны...' : 'Справочник салонов'}
        columns={columns}
        rows={salons}
        emptyText="Салоны пока не созданы или недоступны для текущей роли."
      />

      {selectedSalon ? (
        <DataTable
          caption={`Мастера салона: ${selectedSalon.name}`}
          columns={masterColumns}
          rows={masters}
          emptyText="Для этого салона пока нет закреплённых мастеров."
        />
      ) : null}
    </section>
  )
}
