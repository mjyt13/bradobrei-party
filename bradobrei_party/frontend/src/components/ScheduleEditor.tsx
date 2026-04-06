type ScheduleDraft = Record<string, string>

const scheduleDays: Array<{ key: string; label: string }> = [
  { key: 'mon', label: 'Понедельник' },
  { key: 'tue', label: 'Вторник' },
  { key: 'wed', label: 'Среда' },
  { key: 'thu', label: 'Четверг' },
  { key: 'fri', label: 'Пятница' },
  { key: 'sat', label: 'Суббота' },
  { key: 'sun', label: 'Воскресенье' },
]

export const defaultScheduleDraft: ScheduleDraft = {
  mon: '10:00-20:00',
  tue: '10:00-20:00',
  wed: '',
  thu: '',
  fri: '',
  sat: '',
  sun: '',
}

export function scheduleDraftFromJson(
  value?: string,
  fallback: ScheduleDraft = defaultScheduleDraft,
) {
  const draft = { ...fallback }
  const raw = (value || '').trim()
  if (!raw) {
    return draft
  }

  try {
    const parsed = JSON.parse(raw) as Record<string, unknown>
    if (!parsed || Array.isArray(parsed) || typeof parsed !== 'object') {
      return draft
    }

    for (const day of scheduleDays) {
      const currentValue = parsed[day.key]
      draft[day.key] = typeof currentValue === 'string' ? currentValue : ''
    }
  } catch {
    return draft
  }

  return draft
}

export function scheduleJsonFromDraft(draft: ScheduleDraft) {
  const payload = Object.fromEntries(
    scheduleDays
      .map((day) => [day.key, (draft[day.key] || '').trim()] as const)
      .filter(([, value]) => value),
  )

  return Object.keys(payload).length ? JSON.stringify(payload) : ''
}

interface ScheduleEditorProps {
  value?: string
  onChange: (nextValue: string) => void
  label: string
  hint?: string
  fallback?: ScheduleDraft
  placeholder?: string
}

export function ScheduleEditor({
  value,
  onChange,
  label,
  hint,
  fallback = defaultScheduleDraft,
  placeholder = '10:00-20:00',
}: ScheduleEditorProps) {
  const draft = scheduleDraftFromJson(value, fallback)

  function updateDay(dayKey: string, nextValue: string) {
    const nextDraft = { ...draft, [dayKey]: nextValue }
    onChange(scheduleJsonFromDraft(nextDraft))
  }

  return (
    <div className="field field-wide">
      <span>{label}</span>
      <div className="card-form-grid">
        {scheduleDays.map((day) => (
          <label key={day.key} className="field">
            <span>{day.label}</span>
            <input
              value={draft[day.key] || ''}
              onChange={(event) => updateDay(day.key, event.target.value)}
              placeholder={placeholder}
            />
          </label>
        ))}
      </div>
      {hint ? <span className="field-hint">{hint}</span> : null}
    </div>
  )
}
