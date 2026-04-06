import { useNavigate } from 'react-router-dom'

export function ErrorPage() {
  const navigate = useNavigate()

  return (
    <section className="page-section">
      <div className="page-header">
        <div>
          <p className="eyebrow">Навигация</p>
          <h2>Страница не найдена</h2>
          <p className="section-description">
            Этот маршрут недоступен или устарел. Вернёмся на главную страницу панели и
            продолжим работу оттуда.
          </p>
        </div>
      </div>

      <div className="button-row">
        <button
          type="button"
          className="primary-button"
          onClick={() => navigate('/', { replace: true })}
        >
          На главную
        </button>
      </div>
    </section>
  )
}
