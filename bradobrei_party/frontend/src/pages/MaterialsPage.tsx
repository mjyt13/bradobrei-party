import { useEffect, useMemo, useState } from 'react'
import { useOutletContext } from 'react-router-dom'
import { getReadableErrorMessage } from '../api/errors'
import { materialService } from '../api/services/materialService'
import { useConfirmDialog } from '../components/ConfirmDialog'
import { DataTable, type TableColumn } from '../components/DataTable'
import { materialUnitOptions } from '../shared/options'
import type { AppShellOutletContext } from '../types/appShell'
import type { MaterialDto, UpsertMaterialRequestDto } from '../types/dto/entities'

const initialForm: UpsertMaterialRequestDto = {
  name: '',
  unit: 'мл',
}

function buildMaterialColumns(
  canManageMaterials: boolean,
  onEdit: (material: MaterialDto) => void,
  onDelete: (material: MaterialDto) => void,
): Array<TableColumn<MaterialDto>> {
  return [
    { key: 'name', header: 'Материал', render: (row) => row.name },
    { key: 'unit', header: 'Единица', render: (row) => row.unit || '—' },
    {
      key: 'actions',
      header: 'Действия',
      render: (row) =>
        canManageMaterials ? (
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

export function MaterialsPage() {
  const { currentUser } = useOutletContext<AppShellOutletContext>()
  const { confirm, dialog } = useConfirmDialog()
  const [form, setForm] = useState(initialForm)
  const [materials, setMaterials] = useState<MaterialDto[]>([])
  const [editingId, setEditingId] = useState<number | null>(null)
  const [loading, setLoading] = useState(true)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')
  const [message, setMessage] = useState('')

  const canManageMaterials = currentUser?.role === 'ADMINISTRATOR'
  const materialColumns = useMemo(
    () => buildMaterialColumns(canManageMaterials, startEdit, handleDelete),
    [canManageMaterials],
  )

  async function loadMaterials() {
    setLoading(true)
    try {
      setMaterials(await materialService.getAll())
    } catch (requestError) {
      setError(getReadableErrorMessage(requestError, 'Не удалось загрузить материалы.'))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void loadMaterials()
  }, [])

  function startEdit(material: MaterialDto) {
    if (!canManageMaterials) {
      return
    }

    setEditingId(material.id)
    setForm({
      name: material.name,
      unit: material.unit,
    })
  }

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (!canManageMaterials) {
      return
    }

    setSubmitting(true)
    setError('')
    setMessage('')

    try {
      if (editingId) {
        await materialService.update(editingId, form)
        setMessage(`Материал #${editingId} обновлён.`)
      } else {
        const created = await materialService.create(form)
        setMessage(`Материал "${created.name}" создан.`)
      }

      setForm(initialForm)
      setEditingId(null)
      await loadMaterials()
    } catch (requestError) {
      setError(getReadableErrorMessage(requestError, 'Не удалось сохранить материал.'))
    } finally {
      setSubmitting(false)
    }
  }

  async function handleDelete(material: MaterialDto) {
    if (!canManageMaterials) {
      return
    }

    const shouldContinue = await confirm({
      title: 'Удаление материала',
      message: `Удалить материал "${material.name}"?`,
      confirmLabel: 'Удалить',
      variant: 'danger',
    })
    if (!shouldContinue) {
      return
    }

    try {
      await materialService.remove(material.id)
      setMessage(`Материал "${material.name}" удалён.`)
      await loadMaterials()
    } catch (requestError) {
      setError(getReadableErrorMessage(requestError, 'Не удалось удалить материал.'))
    }
  }

  return (
    <section className="page-section">
      {dialog}
      <div className="page-header">
        <p className="eyebrow">Склад и расход</p>
        <h2>Материалы</h2>
        <p className="section-description">
          Справочник материалов для каталога расходников и норм списания. Управление доступно
          администратору, остальные роли работают в режиме просмотра.
        </p>
      </div>

      {message ? <div className="alert alert-success">{message}</div> : null}
      {error ? <div className="alert alert-error">{error}</div> : null}

      {canManageMaterials ? (
        <form className="card-form card-form-grid" onSubmit={handleSubmit}>
          <label className="field">
            <span>Название материала</span>
            <input
              value={form.name}
              onChange={(event) => setForm((current) => ({ ...current, name: event.target.value }))}
              placeholder="Шампунь для бороды"
              required
            />
          </label>
          <label className="field">
            <span>Единица измерения</span>
            <select
              value={form.unit}
              onChange={(event) => setForm((current) => ({ ...current, unit: event.target.value }))}
            >
              {materialUnitOptions.map((unit) => (
                <option key={unit} value={unit}>
                  {unit}
                </option>
              ))}
            </select>
          </label>
          <div className="button-row field-wide">
            <button type="submit" className="primary-button" disabled={submitting}>
              {submitting ? 'Сохраняем...' : editingId ? 'Обновить материал' : 'Создать материал'}
            </button>
            {editingId ? (
              <button
                type="button"
                className="ghost-button"
                onClick={() => {
                  setEditingId(null)
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
            Для вашей роли доступен просмотр материалов без изменения справочника.
          </div>
        </div>
      )}

      <DataTable
        caption={loading ? 'Загружаем материалы...' : 'Справочник материалов'}
        columns={materialColumns}
        rows={materials}
        emptyText="Материалы пока отсутствуют."
      />
    </section>
  )
}
