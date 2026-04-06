import { useEffect, useMemo, useState } from 'react'
import { useOutletContext } from 'react-router-dom'
import { getReadableErrorMessage } from '../api/errors'
import { employeeService } from '../api/services/employeeService'
import { salonService } from '../api/services/salonService'
import { useConfirmDialog } from '../components/ConfirmDialog'
import { DataTable, type TableColumn } from '../components/DataTable'
import { ScheduleEditor } from '../components/ScheduleEditor'
import { formatCurrency, formatJsonPreview, formatRole } from '../shared/formatters'
import { employeeRoleOptions } from '../shared/options'
import type { AppShellOutletContext } from '../types/appShell'
import type { EmployeeManagementDto, UpdateEmployeeRequestDto } from '../types/dto/employee'
import type { SalonDto } from '../types/dto/entities'

const initialForm: UpdateEmployeeRequestDto = {
  username: '',
  full_name: '',
  phone: '',
  email: '',
  role: 'ADVANCED_MASTER',
  specialization: '',
  expected_salary: 85000,
  work_schedule: '{"mon":"10:00-19:00","wed":"10:00-19:00"}',
  salon_ids: [],
}

function buildEmployeeColumns(
  canManageEmployees: boolean,
  onEdit: (employee: EmployeeManagementDto) => void,
  onFire: (employee: EmployeeManagementDto) => void,
): Array<TableColumn<EmployeeManagementDto>> {
  return [
    {
      key: 'employee',
      header: 'Сотрудник',
      render: (row) => row.user?.full_name || `Профиль #${row.id}`,
    },
    {
      key: 'username',
      header: 'Логин',
      render: (row) => row.user?.username || '—',
    },
    {
      key: 'role',
      header: 'Роль',
      render: (row) => formatRole(row.user?.role || '—'),
    },
    {
      key: 'salary',
      header: 'Оклад',
      render: (row) => formatCurrency(row.expected_salary),
    },
    {
      key: 'salons',
      header: 'Салоны',
      render: (row) =>
        row.salons?.length
          ? row.salons.map((salon) => `#${salon.id} ${salon.name}`).join(', ')
          : '—',
    },
    {
      key: 'schedule',
      header: 'График',
      render: (row) => formatJsonPreview(row.work_schedule),
    },
    {
      key: 'actions',
      header: 'Действия',
      render: (row) =>
        canManageEmployees ? (
          <div className="table-actions">
            <button type="button" className="ghost-button button-small" onClick={() => onEdit(row)}>
              Изменить
            </button>
            <button type="button" className="danger-button button-small" onClick={() => onFire(row)}>
              Уволить
            </button>
          </div>
        ) : (
          <span className="topbar-meta">Просмотр</span>
        ),
    },
  ]
}

export function EmployeesPage() {
  const { currentUser } = useOutletContext<AppShellOutletContext>()
  const { confirm, dialog } = useConfirmDialog()
  const [employees, setEmployees] = useState<EmployeeManagementDto[]>([])
  const [salons, setSalons] = useState<SalonDto[]>([])
  const [editingEmployee, setEditingEmployee] = useState<EmployeeManagementDto | null>(null)
  const [form, setForm] = useState(initialForm)
  const [loading, setLoading] = useState(true)
  const [submitting, setSubmitting] = useState(false)
  const [message, setMessage] = useState('')
  const [error, setError] = useState('')

  const canManageEmployees =
    currentUser?.role === 'ADMINISTRATOR' || currentUser?.role === 'HR_SPECIALIST'

  const employeeColumns = useMemo(
    () => buildEmployeeColumns(canManageEmployees, startEdit, handleFire),
    [canManageEmployees],
  )

  useEffect(() => {
    void loadPageData()
  }, [])

  async function loadPageData() {
    setLoading(true)
    setError('')

    try {
      const [employeesResponse, salonsResponse] = await Promise.all([
        employeeService.getAll(),
        salonService.getAll(),
      ])
      setEmployees(employeesResponse)
      setSalons(salonsResponse)
    } catch (requestError) {
      setError(getReadableErrorMessage(requestError, 'Не удалось загрузить сотрудников.'))
      setEmployees([])
    } finally {
      setLoading(false)
    }
  }

  function toggleSalon(salonId: number) {
    setForm((current) => ({
      ...current,
      salon_ids: current.salon_ids.includes(salonId)
        ? current.salon_ids.filter((id) => id !== salonId)
        : [...current.salon_ids, salonId],
    }))
  }

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (!editingEmployee || !canManageEmployees) {
      return
    }

    setSubmitting(true)
    setMessage('')
    setError('')

    try {
      await employeeService.update(editingEmployee.id, {
        ...form,
        expected_salary: Number(form.expected_salary),
      })
      setMessage(`Данные сотрудника #${editingEmployee.id} обновлены.`)
      setEditingEmployee(null)
      setForm(initialForm)
      await loadPageData()
    } catch (requestError) {
      setError(getReadableErrorMessage(requestError, 'Не удалось обновить сотрудника.'))
    } finally {
      setSubmitting(false)
    }
  }

  async function handleFire(employee: EmployeeManagementDto) {
    if (!canManageEmployees) {
      return
    }

    const shouldContinue = await confirm({
      title: 'Увольнение сотрудника',
      message: `Уволить сотрудника "${employee.user?.full_name || employee.id}"?`,
      confirmLabel: 'Уволить',
      variant: 'danger',
    })
    if (!shouldContinue) {
      return
    }

    setMessage('')
    setError('')

    try {
      await employeeService.fire(employee.id)
      setMessage(`Сотрудник "${employee.user?.full_name || employee.id}" уволен.`)
      if (editingEmployee?.id === employee.id) {
        setEditingEmployee(null)
        setForm(initialForm)
      }
      await loadPageData()
    } catch (requestError) {
      setError(getReadableErrorMessage(requestError, 'Не удалось уволить сотрудника.'))
    }
  }

  function startEdit(employee: EmployeeManagementDto) {
    if (!canManageEmployees) {
      return
    }

    setEditingEmployee(employee)
    setForm({
      username: employee.user?.username || '',
      full_name: employee.user?.full_name || '',
      phone: employee.user?.phone || '',
      email: employee.user?.email || '',
      role: employee.user?.role || 'ADVANCED_MASTER',
      specialization: employee.specialization || '',
      expected_salary: employee.expected_salary,
      work_schedule: employee.work_schedule || '',
      salon_ids: employee.salons?.map((salon) => salon.id) || [],
    })
    setMessage('')
    setError('')
  }

  return (
    <section className="page-section">
      {dialog}
      <div className="page-header">
        <p className="eyebrow">Кадровый контур</p>
        <h2>Управление сотрудниками</h2>
        <p className="section-description">
          Редактирование профилей, ролей, графика и закрепления за салонами. Для управляющего
          сетью экран работает в режиме просмотра, а редактирование доступно администратору и HR.
        </p>
      </div>

      {message ? <div className="alert alert-success">{message}</div> : null}
      {error ? <div className="alert alert-error">{error}</div> : null}

      {editingEmployee && canManageEmployees ? (
        <form className="card-form card-form-grid" onSubmit={handleSubmit}>
          <label className="field">
            <span>Логин</span>
            <input
              value={form.username}
              onChange={(event) => setForm((current) => ({ ...current, username: event.target.value }))}
              required
            />
          </label>
          <label className="field">
            <span>Роль</span>
            <select
              value={form.role}
              onChange={(event) =>
                setForm((current) => ({
                  ...current,
                  role: event.target.value as UpdateEmployeeRequestDto['role'],
                }))
              }
            >
              {employeeRoleOptions.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
          </label>
          <label className="field field-wide">
            <span>ФИО</span>
            <input
              value={form.full_name}
              onChange={(event) => setForm((current) => ({ ...current, full_name: event.target.value }))}
              required
            />
          </label>
          <label className="field">
            <span>Телефон</span>
            <input
              value={form.phone}
              onChange={(event) => setForm((current) => ({ ...current, phone: event.target.value }))}
            />
          </label>
          <label className="field">
            <span>Email</span>
            <input
              type="email"
              value={form.email}
              onChange={(event) => setForm((current) => ({ ...current, email: event.target.value }))}
            />
          </label>
          <label className="field">
            <span>Оклад</span>
            <input
              type="number"
              min="0"
              value={form.expected_salary}
              onChange={(event) =>
                setForm((current) => ({ ...current, expected_salary: Number(event.target.value) }))
              }
            />
          </label>
          <label className="field field-wide">
            <span>Специализация</span>
            <input
              value={form.specialization}
              onChange={(event) => setForm((current) => ({ ...current, specialization: event.target.value }))}
            />
          </label>

          <ScheduleEditor
            label="График работы"
            value={form.work_schedule}
            onChange={(nextValue) => setForm((current) => ({ ...current, work_schedule: nextValue }))}
            hint="Оставьте день пустым, если сотрудник в этот день не работает."
            placeholder="10:00-19:00"
          />

          <div className="field field-wide">
            <span>Закрепление за салонами</span>
            <div className="selection-grid">
              {salons.map((salon) => (
                <label key={salon.id} className="selection-item">
                  <input
                    type="checkbox"
                    checked={form.salon_ids.includes(salon.id)}
                    onChange={() => toggleSalon(salon.id)}
                  />
                  <span>
                    #{salon.id} {salon.name}
                  </span>
                </label>
              ))}
            </div>
            <span className="field-hint">
              ID остаётся видимым рядом с названием, чтобы было удобно сверять его с API и
              отчётами.
            </span>
          </div>
          <div className="button-row field-wide">
            <button type="submit" className="primary-button" disabled={submitting}>
              {submitting ? 'Сохраняем...' : 'Сохранить изменения'}
            </button>
            <button
              type="button"
              className="ghost-button"
              onClick={() => {
                setEditingEmployee(null)
                setForm(initialForm)
              }}
            >
              Отменить
            </button>
          </div>
        </form>
      ) : null}

      {!canManageEmployees ? (
        <div className="filter-card">
          <div className="report-note">
            Для вашей роли доступен просмотр сотрудников без изменения кадровых данных.
          </div>
        </div>
      ) : null}

      <DataTable
        caption={loading ? 'Загружаем сотрудников...' : 'Список сотрудников'}
        columns={employeeColumns}
        rows={employees}
        emptyText="Сотрудники пока не найдены."
      />
    </section>
  )
}
