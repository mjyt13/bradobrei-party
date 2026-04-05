import { useEffect, useState } from 'react'
import { ApiError } from '../api/client'
import { useConfirmDialog } from '../components/ConfirmDialog'
import { DataTable, type TableColumn } from '../components/DataTable'
import { SalonMap } from '../components/SalonMap'
import { salonService } from '../api/services/salonService'
import type { EmployeeProfileSummaryDto, SalonDto, UpsertSalonRequestDto } from '../types/dto/entities'
import { formatCurrency, formatJsonPreview, formatRole, formatSalonStatus } from '../shared/formatters'
import { salonStatusOptions } from '../shared/options'

type WorkingHoursDraft = Record<string, string>

const workingDayLabels: Array<{ key: string; label: string }> = [
  { key: 'mon', label: 'Понедельник' },
  { key: 'tue', label: 'Вторник' },
  { key: 'wed', label: 'Среда' },
  { key: 'thu', label: 'Четверг' },
  { key: 'fri', label: 'Пятница' },
  { key: 'sat', label: 'Суббота' },
  { key: 'sun', label: 'Воскресенье' },
]

const initialWorkingHoursDraft: WorkingHoursDraft = {
  mon: '10:00-20:00',
  tue: '10:00-20:00',
  wed: '',
  thu: '',
  fri: '',
  sat: '',
  sun: '',
}

const initialForm: UpsertSalonRequestDto = {
  name: '',
  address: '',
  location: '',
  working_hours: '{"mon":"10:00-20:00","tue":"10:00-20:00"}',
  status: 'OPEN',
  max_staff: 8,
  base_hourly_rate: 1400,
}

function draftFromWorkingHours(value?: string): WorkingHoursDraft {
  const draft = { ...initialWorkingHoursDraft }
  const raw = (value || '').trim()
  if (!raw) {
    return draft
  }

  try {
    const parsed = JSON.parse(raw) as Record<string, unknown>
    if (!parsed || Array.isArray(parsed) || typeof parsed !== 'object') {
      return draft
    }

    for (const day of workingDayLabels) {
      const currentValue = parsed[day.key]
      draft[day.key] = typeof currentValue === 'string' ? currentValue : ''
    }
  } catch {
    return draft
  }

  return draft
}

function workingHoursToJson(draft: WorkingHoursDraft) {
  const payload = Object.fromEntries(
    workingDayLabels
      .map((day) => [day.key, (draft[day.key] || '').trim()] as const)
      .filter(([, value]) => value),
  )

  return Object.keys(payload).length ? JSON.stringify(payload) : ''
}

function syncFormWithDraft(current: UpsertSalonRequestDto, draft: WorkingHoursDraft): UpsertSalonRequestDto {
  return {
    ...current,
    working_hours: workingHoursToJson(draft),
  }
}

const salonColumns = (
  onEdit: (salon: SalonDto) => void,
  onDelete: (salon: SalonDto) => void,
  onShowMasters: (salon: SalonDto) => void,
): Array<TableColumn<SalonDto>> => [
  { key: 'name', header: 'Салон', render: (row) => row.name },
  { key: 'address', header: 'Адрес', render: (row) => row.address },
  {
    key: 'status',
    header: 'Статус',
    render: (row) => <span className="status-pill">{formatSalonStatus(row.status)}</span>,
  },
  { key: 'staff', header: 'Штат', render: (row) => `${row.max_staff} чел.` },
  { key: 'rate', header: 'Базовая ставка', render: (row) => formatCurrency(row.base_hourly_rate) },
  { key: 'hours', header: 'Часы', render: (row) => formatJsonPreview(row.working_hours) },
  {
    key: 'actions',
    header: 'Действия',
    render: (row) => (
      <div className="table-actions">
        <button type="button" className="ghost-button button-small" onClick={() => onEdit(row)}>
          Изменить
        </button>
        <button type="button" className="ghost-button button-small" onClick={() => onShowMasters(row)}>
          Мастера
        </button>
        <button type="button" className="danger-button button-small" onClick={() => onDelete(row)}>
          Удалить
        </button>
      </div>
    ),
  },
]

const masterColumns: Array<TableColumn<EmployeeProfileSummaryDto>> = [
  { key: 'name', header: 'Сотрудник', render: (row) => row.user?.full_name || `Профиль #${row.id}` },
  { key: 'role', header: 'Роль', render: (row) => formatRole(row.user?.role || '—') },
  { key: 'spec', header: 'Специализация', render: (row) => row.specialization || '—' },
  { key: 'services', header: 'Услуги', render: (row) => row.services?.map((service) => service.name).join(', ') || '—' },
]

export function SalonsPage() {
  const { confirm, dialog } = useConfirmDialog()
  const [form, setForm] = useState(initialForm)
  const [workingHoursDraft, setWorkingHoursDraft] = useState<WorkingHoursDraft>(draftFromWorkingHours(initialForm.working_hours))
  const [salons, setSalons] = useState<SalonDto[]>([])
  const [masters, setMasters] = useState<EmployeeProfileSummaryDto[]>([])
  const [selectedSalon, setSelectedSalon] = useState<SalonDto | null>(null)
  const [editingId, setEditingId] = useState<number | null>(null)
  const [loading, setLoading] = useState(true)
  const [submitting, setSubmitting] = useState(false)
  const [geocodeSubmitting, setGeocodeSubmitting] = useState(false)
  const [error, setError] = useState('')
  const [message, setMessage] = useState('')

  async function loadSalons() {
    setLoading(true)
    try {
      setSalons(await salonService.getAll())
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : 'Не удалось загрузить салоны.')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void loadSalons()
  }, [])

  function updateWorkingHoursDay(dayKey: string, value: string) {
    setWorkingHoursDraft((current) => {
      const next = { ...current, [dayKey]: value }
      setForm((formCurrent) => syncFormWithDraft(formCurrent, next))
      return next
    })
  }

  function resetForm() {
    setEditingId(null)
    setForm(initialForm)
    setWorkingHoursDraft(draftFromWorkingHours(initialForm.working_hours))
  }

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setSubmitting(true)
    setError('')
    setMessage('')

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
      setError(requestError instanceof Error ? requestError.message : 'Не удалось сохранить салон.')
    } finally {
      setSubmitting(false)
    }
  }

  async function handleDelete(salon: SalonDto) {
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
      setError(requestError instanceof Error ? requestError.message : 'Не удалось удалить салон.')
    }
  }

  async function handleGeocodeAddress() {
    setGeocodeSubmitting(true)
    setError('')
    setMessage('')
    try {
      const res = await salonService.geocodeAddress(form.address)
      setForm((current) => ({
        ...current,
        location: `${res.latitude}, ${res.longitude}`,
      }))
      setMessage(`Адрес подтверждён (${res.provider}): ${res.formatted_address}`)
    } catch (requestError) {
      if (requestError instanceof ApiError && requestError.status === 503) {
        setError('На сервере не включён геокодер. Задайте GEOCODER_PROVIDER и ключи в backend .env или используйте ручные координаты.')
      } else {
        setError(requestError instanceof Error ? requestError.message : 'Не удалось геокодировать адрес.')
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
      setError(requestError instanceof Error ? requestError.message : 'Не удалось загрузить мастеров салона.')
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
          Здесь можно вести справочник филиалов, обновлять параметры работы и быстро смотреть мастеров по выбранному салону.
        </p>
      </div>

      <SalonMap salons={salons} />

      {message ? <div className="alert alert-success">{message}</div> : null}
      {error ? <div className="alert alert-error">{error}</div> : null}

      <form className="card-form card-form-grid" onSubmit={handleSubmit}>
        <label className="field">
          <span>Название салона</span>
          <input value={form.name} onChange={(event) => setForm((current) => ({ ...current, name: event.target.value }))} placeholder="Bradobrei Party Center" required />
        </label>
        <label className="field">
          <span>Статус</span>
          <select value={form.status} onChange={(event) => setForm((current) => ({ ...current, status: event.target.value as UpsertSalonRequestDto['status'] }))}>
            {salonStatusOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </label>
        <label className="field field-wide">
          <span>Адрес</span>
          <input value={form.address} onChange={(event) => setForm((current) => ({ ...current, address: event.target.value }))} placeholder="Екатеринбург, ул. Малышева, 12" required />
        </label>
        <label className="field field-wide">
          <span>Координаты</span>
          <div className="field-inline">
            <input
              value={form.location || ''}
              onChange={(event) => setForm((current) => ({ ...current, location: event.target.value }))}
              placeholder="58.0141, 56.2230 (если не настроен серверный геокодер)"
            />
            <button type="button" className="ghost-button" onClick={handleGeocodeAddress} disabled={geocodeSubmitting || !form.address.trim()}>
              {geocodeSubmitting ? 'Проверка...' : 'Проверить адрес на сервере'}
            </button>
          </div>
          <span className="field-hint">При включённом серверном геокодере координаты при сохранении будут уточняться по адресу.</span>
        </label>
        <label className="field">
          <span>Макс. персонал</span>
          <input type="number" min="1" value={form.max_staff} onChange={(event) => setForm((current) => ({ ...current, max_staff: Number(event.target.value) }))} />
        </label>
        <label className="field">
          <span>Базовая ставка</span>
          <input type="number" min="0" value={form.base_hourly_rate} onChange={(event) => setForm((current) => ({ ...current, base_hourly_rate: Number(event.target.value) }))} />
        </label>
        <div className="field field-wide">
          <span>Часы работы</span>
          <div className="card-form-grid">
            {workingDayLabels.map((day) => (
              <label key={day.key} className="field">
                <span>{day.label}</span>
                <input
                  value={workingHoursDraft[day.key] || ''}
                  onChange={(event) => updateWorkingHoursDay(day.key, event.target.value)}
                  placeholder="10:00-20:00"
                />
              </label>
            ))}
          </div>
          <span className="field-hint">Оставьте день пустым, если салон в этот день не работает.</span>
        </div>
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

      <DataTable
        caption={loading ? 'Загружаем салоны...' : 'Справочник салонов'}
        columns={salonColumns(
          (salon) => {
            const nextForm = {
              name: salon.name,
              address: salon.address,
              location: salon.location || '',
              working_hours: salon.working_hours || '',
              status: salon.status,
              max_staff: salon.max_staff,
              base_hourly_rate: salon.base_hourly_rate,
            }
            setEditingId(salon.id)
            setForm(nextForm)
            setWorkingHoursDraft(draftFromWorkingHours(nextForm.working_hours))
          },
          handleDelete,
          handleShowMasters,
        )}
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
