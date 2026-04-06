import { useEffect, useMemo, useState } from 'react'
import { useOutletContext } from 'react-router-dom'
import { getReadableErrorMessage } from '../api/errors'
import { employeeService } from '../api/services/employeeService'
import { materialService } from '../api/services/materialService'
import { serviceService } from '../api/services/serviceService'
import { useConfirmDialog } from '../components/ConfirmDialog'
import { DataTable, type TableColumn } from '../components/DataTable'
import { formatCurrency, formatRole } from '../shared/formatters'
import type { AppShellOutletContext } from '../types/appShell'
import type { EmployeeManagementDto } from '../types/dto/employee'
import type { MaterialDto, ServiceDto, UpsertServiceRequestDto } from '../types/dto/entities'

interface ServiceFormState extends UpsertServiceRequestDto {
  master_ids: number[]
  materials: Array<{
    material_id: number
    quantity_per_use: number
  }>
}

const initialForm: ServiceFormState = {
  name: '',
  description: '',
  price: 1800,
  duration_minutes: 60,
  master_ids: [],
  materials: [],
}

const masterRoles = new Set(['BASIC_MASTER', 'ADVANCED_MASTER'])

function buildServiceColumns(
  canManageServices: boolean,
  onEdit: (service: ServiceDto) => void,
  onDelete: (service: ServiceDto) => void,
): Array<TableColumn<ServiceDto>> {
  return [
    { key: 'name', header: 'Услуга', render: (row) => row.name },
    { key: 'description', header: 'Описание', render: (row) => row.description || '—' },
    { key: 'price', header: 'Цена', render: (row) => formatCurrency(row.price) },
    { key: 'duration', header: 'Длительность', render: (row) => `${row.duration_minutes} мин.` },
    {
      key: 'masters',
      header: 'Мастера',
      render: (row) =>
        row.employees?.length
          ? row.employees
              .map((employee) => employee.user?.full_name || `Профиль #${employee.id}`)
              .join(', ')
          : '—',
    },
    {
      key: 'materials',
      header: 'Материалы',
      render: (row) =>
        row.materials?.length
          ? row.materials
              .map((item) => `${item.material?.name || item.material_id} x${item.quantity_per_use}`)
              .join(', ')
          : '—',
    },
    {
      key: 'actions',
      header: 'Действия',
      render: (row) =>
        canManageServices ? (
          <div className="table-actions">
            <button type="button" className="ghost-button button-small" onClick={() => onEdit(row)}>
              Изменить
            </button>
            <button type="button" className="danger-button button-small" onClick={() => onDelete(row)}>
              Удалить
            </button>
          </div>
        ) : (
          <span className="topbar-meta">Просмотр</span>
        ),
    },
  ]
}

export function ServicesPage() {
  const { currentUser } = useOutletContext<AppShellOutletContext>()
  const { confirm, dialog } = useConfirmDialog()
  const [form, setForm] = useState<ServiceFormState>(initialForm)
  const [services, setServices] = useState<ServiceDto[]>([])
  const [materials, setMaterials] = useState<MaterialDto[]>([])
  const [employees, setEmployees] = useState<EmployeeManagementDto[]>([])
  const [editingService, setEditingService] = useState<ServiceDto | null>(null)
  const [loading, setLoading] = useState(true)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')
  const [message, setMessage] = useState('')

  const canManageServices =
    currentUser?.role === 'ADMINISTRATOR' || currentUser?.role === 'ADVANCED_MASTER'

  const masters = useMemo(
    () => employees.filter((employee) => employee.user && masterRoles.has(employee.user.role)),
    [employees],
  )

  const serviceColumns = useMemo(
    () => buildServiceColumns(canManageServices, startEdit, handleDelete),
    [canManageServices],
  )

  async function loadPageData() {
    setLoading(true)
    setError('')

    try {
      const [servicesData, materialsData] = await Promise.all([
        serviceService.getAll(),
        materialService.getAll(),
      ])

      let employeesData: EmployeeManagementDto[] = []
      if (
        currentUser?.role === 'ADMINISTRATOR' ||
        currentUser?.role === 'HR_SPECIALIST' ||
        currentUser?.role === 'NETWORK_MANAGER'
      ) {
        employeesData = await employeeService.getAll()
      }

      setServices(servicesData)
      setMaterials(materialsData)
      setEmployees(employeesData)
    } catch (requestError) {
      setError(getReadableErrorMessage(requestError, 'Не удалось загрузить справочники услуг.'))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void loadPageData()
  }, [currentUser?.role])

  function toggleMaster(userId: number) {
    setForm((current) => ({
      ...current,
      master_ids: current.master_ids.includes(userId)
        ? current.master_ids.filter((id) => id !== userId)
        : [...current.master_ids, userId],
    }))
  }

  function updateMaterialSelection(materialId: number, enabled: boolean) {
    setForm((current) => {
      if (!enabled) {
        return {
          ...current,
          materials: current.materials.filter((item) => item.material_id !== materialId),
        }
      }

      if (current.materials.some((item) => item.material_id === materialId)) {
        return current
      }

      return {
        ...current,
        materials: [...current.materials, { material_id: materialId, quantity_per_use: 1 }],
      }
    })
  }

  function updateMaterialQuantity(materialId: number, quantity: number) {
    setForm((current) => ({
      ...current,
      materials: current.materials.map((item) =>
        item.material_id === materialId
          ? { ...item, quantity_per_use: Number.isFinite(quantity) && quantity > 0 ? quantity : 0 }
          : item,
      ),
    }))
  }

  async function syncRelations(serviceId: number, existing?: ServiceDto | null) {
    const materialPayload = form.materials
      .filter((item) => item.quantity_per_use > 0)
      .map((item) => ({
        material_id: item.material_id,
        quantity_per_use: Number(item.quantity_per_use),
      }))

    await serviceService.setMaterials(serviceId, materialPayload)

    const currentAssignments = existing?.employees ?? []
    const currentUserIds = new Set(currentAssignments.map((employee) => employee.user_id))
    const selectedUserIds = new Set(form.master_ids)

    for (const assignment of currentAssignments) {
      if (!selectedUserIds.has(assignment.user_id)) {
        await serviceService.removeMaster(serviceId, assignment.id)
      }
    }

    for (const userId of form.master_ids) {
      if (!currentUserIds.has(userId)) {
        await serviceService.assignMaster(serviceId, userId)
      }
    }
  }

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (!canManageServices) {
      return
    }

    setSubmitting(true)
    setError('')
    setMessage('')

    try {
      const payload: UpsertServiceRequestDto = {
        name: form.name,
        description: form.description,
        price: Number(form.price),
        duration_minutes: Number(form.duration_minutes),
      }

      if (editingService) {
        const updated = await serviceService.update(editingService.id, payload)
        await syncRelations(updated.id, editingService)
        setMessage(`Услуга "${updated.name}" обновлена.`)
      } else {
        const created = await serviceService.create(payload)
        await syncRelations(created.id, null)
        setMessage(`Услуга "${created.name}" создана.`)
      }

      setForm(initialForm)
      setEditingService(null)
      await loadPageData()
    } catch (requestError) {
      setError(getReadableErrorMessage(requestError, 'Не удалось сохранить услугу.'))
    } finally {
      setSubmitting(false)
    }
  }

  async function handleDelete(service: ServiceDto) {
    if (!canManageServices) {
      return
    }

    const shouldContinue = await confirm({
      title: 'Удаление услуги',
      message: `Удалить услугу "${service.name}"?`,
      confirmLabel: 'Удалить',
      variant: 'danger',
    })
    if (!shouldContinue) {
      return
    }

    try {
      await serviceService.remove(service.id)
      setMessage(`Услуга "${service.name}" удалена.`)
      await loadPageData()
    } catch (requestError) {
      setError(getReadableErrorMessage(requestError, 'Не удалось удалить услугу.'))
    }
  }

  function startEdit(service: ServiceDto) {
    if (!canManageServices) {
      return
    }

    setEditingService(service)
    setForm({
      name: service.name,
      description: service.description,
      price: service.price,
      duration_minutes: service.duration_minutes,
      master_ids: service.employees?.map((employee) => employee.user_id) ?? [],
      materials:
        service.materials?.map((item) => ({
          material_id: item.material_id,
          quantity_per_use: item.quantity_per_use,
        })) ?? [],
    })
  }

  return (
    <section className="page-section">
      {dialog}
      <div className="page-header">
        <p className="eyebrow">Прайс и выполнение</p>
        <h2>Услуги</h2>
        <p className="section-description">
          Каталог услуг, связей с мастерами и материалов. Изменение услуг открыто
          администратору и продвинутому мастеру, остальным ролям доступен просмотр.
        </p>
      </div>

      {message ? <div className="alert alert-success">{message}</div> : null}
      {error ? <div className="alert alert-error">{error}</div> : null}

      {canManageServices ? (
        <form className="card-form card-form-grid" onSubmit={handleSubmit}>
          <label className="field">
            <span>Название услуги</span>
            <input
              value={form.name}
              onChange={(event) => setForm((current) => ({ ...current, name: event.target.value }))}
              placeholder="Мужская стрижка"
              required
            />
          </label>
          <label className="field">
            <span>Цена</span>
            <input
              type="number"
              min="0"
              value={form.price}
              onChange={(event) => setForm((current) => ({ ...current, price: Number(event.target.value) }))}
            />
          </label>
          <label className="field">
            <span>Длительность, минут</span>
            <input
              type="number"
              min="60"
              step="5"
              value={form.duration_minutes}
              onChange={(event) =>
                setForm((current) => ({ ...current, duration_minutes: Number(event.target.value) }))
              }
            />
          </label>
          <label className="field field-wide">
            <span>Описание</span>
            <textarea
              rows={4}
              value={form.description}
              onChange={(event) => setForm((current) => ({ ...current, description: event.target.value }))}
              placeholder="Стрижка с мытьём головы и укладкой."
            />
          </label>

          <div className="field field-wide">
            <span>Мастера</span>
            <div className="selection-grid">
              {masters.map((master) => (
                <label key={master.id} className="selection-item">
                  <input
                    type="checkbox"
                    checked={form.master_ids.includes(master.user_id)}
                    onChange={() => toggleMaster(master.user_id)}
                  />
                  <span>
                    {master.user?.full_name || `Профиль #${master.id}`} ({formatRole(master.user?.role || '')})
                  </span>
                </label>
              ))}
              {masters.length === 0 ? (
                <p className="report-note">
                  Доступные мастера для привязки сейчас не загружены для вашей роли.
                </p>
              ) : null}
            </div>
          </div>

          <div className="field field-wide">
            <span>Материалы</span>
            <div className="selection-grid">
              {materials.map((material) => {
                const selected = form.materials.find((item) => item.material_id === material.id)

                return (
                  <div key={material.id} className="selection-item selection-item-row">
                    <label>
                      <input
                        type="checkbox"
                        checked={Boolean(selected)}
                        onChange={(event) => updateMaterialSelection(material.id, event.target.checked)}
                      />
                      <span>
                        {material.name} ({material.unit})
                      </span>
                    </label>
                    <input
                      type="number"
                      min="0"
                      step="0.1"
                      disabled={!selected}
                      value={selected?.quantity_per_use ?? 0}
                      onChange={(event) => updateMaterialQuantity(material.id, Number(event.target.value))}
                    />
                  </div>
                )
              })}
              {materials.length === 0 ? (
                <p className="report-note">Сначала добавьте материалы в справочник.</p>
              ) : null}
            </div>
          </div>

          <div className="button-row field-wide">
            <button type="submit" className="primary-button" disabled={submitting}>
              {submitting ? 'Сохраняем...' : editingService ? 'Обновить услугу' : 'Создать услугу'}
            </button>
            {editingService ? (
              <button
                type="button"
                className="ghost-button"
                onClick={() => {
                  setEditingService(null)
                  setForm(initialForm)
                }}
              >
                Сбросить редактирование
              </button>
            ) : null}
          </div>
        </form>
      ) : (
        <div className="filter-card">
          <div className="report-note">
            Для вашей роли доступен просмотр каталога услуг без изменения справочника.
          </div>
        </div>
      )}

      <DataTable
        caption={loading ? 'Загружаем услуги...' : 'Справочник услуг'}
        columns={serviceColumns}
        rows={services}
        emptyText="Услуги пока отсутствуют."
      />
    </section>
  )
}
