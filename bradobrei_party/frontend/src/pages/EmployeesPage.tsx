import { useEffect, useState } from 'react'
import { useConfirmDialog } from '../components/ConfirmDialog'
import { DataTable, type TableColumn } from '../components/DataTable'
import { employeeService } from '../api/services/employeeService'
import type { EmployeeManagementDto, UpdateEmployeeRequestDto } from '../types/dto/employee'
import { formatCurrency, formatJsonPreview, formatRole } from '../shared/formatters'
import { employeeRoleOptions } from '../shared/options'

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

const employeeColumns = (
  onEdit: (employee: EmployeeManagementDto) => void,
  onFire: (employee: EmployeeManagementDto) => void,
): Array<TableColumn<EmployeeManagementDto>> => [
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
    render: (row) => row.salons?.map((salon) => salon.name).join(', ') || '—',
  },
  {
    key: 'schedule',
    header: 'График',
    render: (row) => formatJsonPreview(row.work_schedule),
  },
  {
    key: 'actions',
    header: 'Действия',
    render: (row) => (
      <div className="table-actions">
        <button type="button" className="ghost-button button-small" onClick={() => onEdit(row)}>
          Изменить
        </button>
        <button type="button" className="danger-button button-small" onClick={() => onFire(row)}>
          Уволить
        </button>
      </div>
    ),
  },
]

function parseSalonIds(value: string) {
  return value
    .split(',')
    .map((item) => item.trim())
    .filter(Boolean)
    .map((item) => Number(item))
    .filter((item) => !Number.isNaN(item) && item > 0)
}

export function EmployeesPage() {
  const { confirm, dialog } = useConfirmDialog()
  const [employees, setEmployees] = useState<EmployeeManagementDto[]>([])
  const [editingEmployee, setEditingEmployee] = useState<EmployeeManagementDto | null>(null)
  const [form, setForm] = useState(initialForm)
  const [salonIdsInput, setSalonIdsInput] = useState('')
  const [loading, setLoading] = useState(true)
  const [submitting, setSubmitting] = useState(false)
  const [message, setMessage] = useState('')
  const [error, setError] = useState('')

  async function loadEmployees() {
    setLoading(true)
    try {
      setEmployees(await employeeService.getAll())
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : 'Не удалось загрузить сотрудников.')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void loadEmployees()
  }, [])

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (!editingEmployee) {
      return
    }

    setSubmitting(true)
    setMessage('')
    setError('')

    try {
      await employeeService.update(editingEmployee.id, {
        ...form,
        expected_salary: Number(form.expected_salary),
        salon_ids: parseSalonIds(salonIdsInput),
      })
      setMessage(`Данные сотрудника #${editingEmployee.id} обновлены.`)
      setEditingEmployee(null)
      setForm(initialForm)
      setSalonIdsInput('')
      await loadEmployees()
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : 'Не удалось обновить сотрудника.')
    } finally {
      setSubmitting(false)
    }
  }

  async function handleFire(employee: EmployeeManagementDto) {
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
        setSalonIdsInput('')
      }
      await loadEmployees()
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : 'Не удалось уволить сотрудника.')
    }
  }

  function startEdit(employee: EmployeeManagementDto) {
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
    setSalonIdsInput((employee.salons || []).map((salon) => salon.id).join(', '))
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
          Редактирование профилей, ролей, графика и закрепления за салонами; увольнение. Найм остаётся на отдельной странице.
        </p>
      </div>

      {message ? <div className="alert alert-success">{message}</div> : null}
      {error ? <div className="alert alert-error">{error}</div> : null}

      {editingEmployee ? (
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
                setForm((current) => ({ ...current, role: event.target.value as UpdateEmployeeRequestDto['role'] }))
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
          <label className="field">
            <span>ID салонов через запятую</span>
            <input
              value={salonIdsInput}
              onChange={(event) => setSalonIdsInput(event.target.value)}
              placeholder="1, 2"
            />
          </label>
          <label className="field field-wide">
            <span>Специализация</span>
            <input
              value={form.specialization}
              onChange={(event) => setForm((current) => ({ ...current, specialization: event.target.value }))}
            />
          </label>
          <label className="field field-wide">
            <span>График JSON</span>
            <textarea
              rows={4}
              value={form.work_schedule}
              onChange={(event) => setForm((current) => ({ ...current, work_schedule: event.target.value }))}
            />
          </label>
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
                setSalonIdsInput('')
              }}
            >
              Отменить
            </button>
          </div>
        </form>
      ) : null}

      <DataTable
        caption={loading ? 'Загружаем сотрудников...' : 'Список сотрудников'}
        columns={employeeColumns(startEdit, handleFire)}
        rows={employees}
        emptyText="Сотрудники пока не найдены."
      />
    </section>
  )
}
