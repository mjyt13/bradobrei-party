import { startTransition, useEffect, useMemo, useRef, useState } from 'react'
import { NavLink, Outlet, useLocation, useNavigate } from 'react-router-dom'
import { apiConfig } from '../api/config'
import { getReadableErrorMessage } from '../api/errors'
import { authService } from '../api/services/authService'
import { tokenStorage } from '../api/services/tokenStorage'
import { formatRole } from '../shared/formatters'
import { canAccessRoute, getDefaultRouteForRole, getVisibleNavigation } from '../shared/rbac'
import type { UserDto } from '../types/dto/auth'

export function AppShell() {
  const [currentUser, setCurrentUser] = useState<UserDto | null>(null)
  const [error, setError] = useState('')
  const navigate = useNavigate()
  const location = useLocation()
  const previousPathRef = useRef(location.pathname)

  const navigation = useMemo(
    () => getVisibleNavigation(currentUser?.role),
    [currentUser?.role],
  )

  useEffect(() => {
    if (previousPathRef.current !== location.pathname) {
      previousPathRef.current = location.pathname
      setError('')
    }
  }, [location.pathname])

  useEffect(() => {
    let cancelled = false

    authService
      .getMe()
      .then((response) => {
        if (!cancelled) {
          setCurrentUser(response.user)
        }
      })
      .catch((requestError: Error) => {
        if (!cancelled) {
          setError(getReadableErrorMessage(requestError, 'Не удалось загрузить профиль текущего пользователя.'))
          tokenStorage.clear()
          startTransition(() => {
            navigate('/auth', { replace: true })
          })
        }
      })

    return () => {
      cancelled = true
    }
  }, [navigate])

  useEffect(() => {
    function handleUnauthorized() {
      setError('Сессия истекла. Выполните вход повторно.')
      startTransition(() => {
        navigate('/auth', { replace: true })
      })
    }

    window.addEventListener('auth:unauthorized', handleUnauthorized)
    return () => window.removeEventListener('auth:unauthorized', handleUnauthorized)
  }, [navigate])

  useEffect(() => {
    if (!currentUser) {
      return
    }

    if (!canAccessRoute(currentUser.role, location.pathname)) {
      setError('Этот раздел недоступен для вашей роли. Открыт ближайший доступный экран.')
      startTransition(() => {
        navigate(getDefaultRouteForRole(currentUser.role), { replace: true })
      })
    }
  }, [currentUser, location.pathname, navigate])

  return (
    <div className="app-shell">
      <aside className="sidebar">
        <div className="brand-block">
          <p className="eyebrow">Bradobrei Party</p>
          <h1>Панель операций и отчётов</h1>
          <p className="lede">
            Система управления и мониторинга работы с салонами, бронированиями, мастерами,
            платежами и аналитикой.
          </p>
        </div>

        <nav className="navigation">
          {navigation.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              className={({ isActive }) => (isActive ? 'nav-link nav-link-active' : 'nav-link')}
            >
              {item.label}
            </NavLink>
          ))}
        </nav>

        <div className="sidebar-footer">
          <a href={apiConfig.docsUrl} target="_blank" rel="noreferrer">
            Swagger backend
          </a>
        </div>
      </aside>

      <main className="content">
        <header className="topbar">
          <div>
            <p className="topbar-label">Текущий пользователь</p>
            {currentUser ? (
              <>
                <strong>{currentUser.full_name}</strong>
                <span className="topbar-meta">{formatRole(currentUser.role)}</span>
              </>
            ) : (
              <span className="topbar-meta">Загрузка профиля...</span>
            )}
          </div>

          <button
            type="button"
            className="ghost-button"
            onClick={() => {
              tokenStorage.clear()
              startTransition(() => {
                navigate('/auth', { replace: true })
              })
            }}
          >
            Выйти
          </button>
        </header>

        {error ? <div className="alert alert-error">{error}</div> : null}
        <Outlet context={{ currentUser }} />
      </main>
    </div>
  )
}
