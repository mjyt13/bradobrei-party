import { useEffect, useState } from 'react'
import { useConfirmDialog } from '../components/ConfirmDialog'
import { DataTable, type TableColumn } from '../components/DataTable'
import { materialService } from '../api/services/materialService'
import type { MaterialDto, UpsertMaterialRequestDto } from '../types/dto/entities'
import { materialUnitOptions } from '../shared/options'

const initialForm: UpsertMaterialRequestDto = {
  name: '',
  unit: 'мл',
}

const materialColumns = (
  onEdit: (material: MaterialDto) => void,
  onDelete: (material: MaterialDto) => void,
): Array<TableColumn<MaterialDto>> => [
  { key: 'name', header: 'Материал', render: (row) => row.name },
  { key: 'unit', header: 'Единица', render: (row) => row.unit || '—' },
  {
    key: 'actions',
    header: 'Действия',
    render: (row) => (
      <div className="table-actions">
        <button type="button" className="ghost-button button-small" onClick={() => onEdit(row)}>
          Изменить
        </button>
        <button type="button" className="danger-button button-small" onClick={() => onDelete(row)}>
          Удалить
        </button>
      </div>
    ),
  },
]

export function MaterialsPage() {
  const { confirm, dialog } = useConfirmDialog()
  const [form, setForm] = useState(initialForm)
  const [materials, setMaterials] = useState<MaterialDto[]>([])
  const [editingId, setEditingId] = useState<number | null>(null)
  const [loading, setLoading] = useState(true)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')
  const [message, setMessage] = useState('')

  async function loadMaterials() {
    setLoading(true)
    try {
      setMaterials(await materialService.getAll())
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : 'Не удалось загрузить материалы.')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void loadMaterials()
  }, [])

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
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
      setError(requestError instanceof Error ? requestError.message : 'Не удалось сохранить материал.')
    } finally {
      setSubmitting(false)
    }
  }

  async function handleDelete(material: MaterialDto) {
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
      setError(requestError instanceof Error ? requestError.message : 'Не удалось удалить материал.')
    }
  }

  return (
    <section className="page-section">
      {dialog}
      <div className="page-header">
        <p className="eyebrow">Склад и расход</p>
        <h2>Материалы</h2>
        <p className="section-description">
          Справочник материалов для пополнения каталога расходников и привязки норм расхода к услугам.
        </p>
      </div>

      {message ? <div className="alert alert-success">{message}</div> : null}
      {error ? <div className="alert alert-error">{error}</div> : null}

      <form className="card-form card-form-grid" onSubmit={handleSubmit}>
        <label className="field">
          <span>Название материала</span>
          <input value={form.name} onChange={(event) => setForm((current) => ({ ...current, name: event.target.value }))} placeholder="Шампунь для бороды" required />
        </label>
        <label className="field">
          <span>Единица измерения</span>
          <select value={form.unit} onChange={(event) => setForm((current) => ({ ...current, unit: event.target.value }))}>
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

      <DataTable
        caption={loading ? 'Загружаем материалы...' : 'Справочник материалов'}
        columns={materialColumns(
          (material) => {
            setEditingId(material.id)
            setForm({
              name: material.name,
              unit: material.unit,
            })
          },
          handleDelete,
        )}
        rows={materials}
        emptyText="Материалы пока отсутствуют."
      />
    </section>
  )
}
