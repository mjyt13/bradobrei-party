import { useEffect, useMemo, useState } from 'react'
import { useOutletContext } from 'react-router-dom'
import { getReadableErrorMessage } from '../api/errors'
import { employeeService } from '../api/services/employeeService'
import { salonService } from '../api/services/salonService'
import { ScheduleEditor } from '../components/ScheduleEditor'
import { employeeRoleOptions } from '../shared/options'
import type { AppShellOutletContext } from '../types/appShell'
import type { SalonDto } from '../types/dto/entities'

const baseInitialForm = {
  username: '',
  password: '',
  full_name: '',
  phone: '',
  email: '',
  role: 'ADVANCED_MASTER' as const,
  specialization: '',
  expected_salary: 65000,
  work_schedule: '{"mon":"10:00-19:00","wed":"10:00-19:00"}',
}

export function EmployeeRegistrationPage() {
  const { currentUser } = useOutletContext<AppShellOutletContext>()
  const [salons, setSalons] = useState<SalonDto[]>([])
  const [form, setForm] = useState({
    ...baseInitialForm,
    salon_id: 0,
  })
  const [message, setMessage] = useState('')
  const [error, setError] = useState('')
  const [submitting, setSubmitting] = useState(false)

  const hasSalonOptions = salons.length > 0

  const selectedSalon = useMemo(
    () => salons.find((salon) => salon.id === form.salon_id) || null,
    [form.salon_id, salons],
  )

  useEffect(() => {
    salonService
      .getAll()
      .then((response) => {
        setSalons(response)
        setForm((current) => ({
          ...current,
          salon_id: response[0]?.id || 0,
        }))
      })
      .catch((requestError) => {
        setError(getReadableErrorMessage(requestError, 'Не удалось загрузить список салонов.'))
      })
  }, [])

  function resetForm() {
    setForm({
      ...baseInitialForm,
      salon_id: salons[0]?.id || 0,
    })
  }

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setSubmitting(true)
    setMessage('')
    setError('')

    try {
      const response = await employeeService.hire({
        ...form,
        expected_salary: Number(form.expected_salary),
        salon_id: Number(form.salon_id),
      })
      setMessage(`Сотрудник создан. Профиль #${response.id} готов к дальнейшей настройке.`)
      resetForm()
    } catch (requestError) {
      setError(getReadableErrorMessage(requestError, 'Не удалось создать сотрудника.'))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <section className="page-section">
      <div className="page-header">
        <div>
          <p className="eyebrow">Кадровый контур</p>
          <h2>Регистрация нового сотрудника</h2>
          <p className="section-description">
            Форма использует DTO backend для `HireEmployeeRequest` и показывает салоны по названию
            и ID, чтобы не выбирать филиал вслепую.
          </p>
          <p className="section-description">
            Текущая роль: {currentUser?.role || 'не определена'}.
          </p>
        </div>
      </div>

      {message ? <div className="alert alert-success">{message}</div> : null}
      {error ? <div className="alert alert-error">{error}</div> : null}

      <form className="card-form card-form-grid" onSubmit={handleSubmit}>
        <label className="field">
          <span>Логин</span>
          <input
            value={form.username}
            onChange={(event) => setForm((current) => ({ ...current, username: event.target.value }))}
            placeholder="master_ivan"
            required
          />
        </label>
        <label className="field">
          <span>Пароль</span>
          <input
            type="password"
            value={form.password}
            onChange={(event) => setForm((current) => ({ ...current, password: event.target.value }))}
            placeholder="password"
            required
          />
        </label>
        <label className="field field-wide">
          <span>ФИО</span>
          <input
            value={form.full_name}
            onChange={(event) => setForm((current) => ({ ...current, full_name: event.target.value }))}
            placeholder="Иван Барбер"
            required
          />
        </label>
        <label className="field">
          <span>Телефон</span>
          <input
            value={form.phone}
            onChange={(event) => setForm((current) => ({ ...current, phone: event.target.value }))}
            placeholder="+79990001122"
            required
          />
        </label>
        <label className="field">
          <span>Email</span>
          <input
            type="email"
            value={form.email}
            onChange={(event) => setForm((current) => ({ ...current, email: event.target.value }))}
            placeholder="ivan.barber@example.com"
            required
          />
        </label>
        <label className="field">
          <span>Роль</span>
          <select
            value={form.role}
            onChange={(event) =>
              setForm((current) => ({ ...current, role: event.target.value as typeof form.role }))
            }
          >
            {employeeRoleOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </label>
        <label className="field">
          <span>Ожидаемый оклад</span>
          <input
            type="number"
            min="0"
            value={form.expected_salary}
            onChange={(event) =>
              setForm((current) => ({
                ...current,
                expected_salary: Number(event.target.value),
              }))
            }
          />
        </label>
        <label className="field field-wide">
          <span>Специализация</span>
          <input
            value={form.specialization}
            onChange={(event) => setForm((current) => ({ ...current, specialization: event.target.value }))}
            placeholder="Fade, beard styling"
          />
        </label>
        <label className="field">
          <span>Салон</span>
          <select
            value={form.salon_id || ''}
            onChange={(event) =>
              setForm((current) => ({ ...current, salon_id: Number(event.target.value) }))
            }
            required
            disabled={!hasSalonOptions}
          >
            {!hasSalonOptions ? <option value="">Салоны не загружены</option> : null}
            {salons.map((salon) => (
              <option key={salon.id} value={salon.id}>
                #{salon.id} {salon.name}
              </option>
            ))}
          </select>
        </label>

        <ScheduleEditor
          label="График работы"
          value={form.work_schedule}
          onChange={(nextValue) => setForm((current) => ({ ...current, work_schedule: nextValue }))}
          hint="Оставьте день пустым, если мастер в этот день не работает."
          placeholder="10:00-19:00"
        />

        <div className="field field-wide">
          <span>Выбранный салон</span>
          {selectedSalon ? (
            <div className="checkbox-grid">
              <label className="checkbox-card">
                <input type="radio" checked readOnly />
                <span>
                  <strong>
                    #{selectedSalon.id} {selectedSalon.name}
                  </strong>
                  <small>{selectedSalon.address}</small>
                </span>
              </label>
            </div>
          ) : (
            <p className="section-description field-hint">Список салонов ещё не загружен.</p>
          )}
        </div>
        <button
          type="submit"
          className="primary-button field-wide"
          disabled={submitting || !hasSalonOptions}
        >
          {submitting ? 'Создаём сотрудника...' : 'Создать сотрудника'}
        </button>
      </form>
    </section>
  )
}
