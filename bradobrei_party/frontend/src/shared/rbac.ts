import type { UserRole } from '../types/dto/common'

export interface NavigationItem {
  to: string
  label: string
  allowedRoles?: UserRole[]
}

export const protectedNavigation: NavigationItem[] = [
  { to: '/salons', label: 'Салоны' },
  {
    to: '/bookings',
    label: 'Бронирования',
    allowedRoles: ['CLIENT', 'BASIC_MASTER', 'ADVANCED_MASTER', 'ACCOUNTANT', 'NETWORK_MANAGER', 'ADMINISTRATOR'],
  },
  { to: '/services', label: 'Услуги' },
  {
    to: '/materials',
    label: 'Материалы',
    allowedRoles: ['ADVANCED_MASTER', 'ACCOUNTANT', 'ADMINISTRATOR'],
  },
  {
    to: '/payments',
    label: 'Платежи',
    allowedRoles: ['ACCOUNTANT', 'NETWORK_MANAGER', 'ADMINISTRATOR'],
  },
  {
    to: '/employees',
    label: 'Сотрудники',
    allowedRoles: ['HR_SPECIALIST', 'NETWORK_MANAGER', 'ADMINISTRATOR'],
  },
  {
    to: '/reports',
    label: 'Отчёты',
    allowedRoles: ['HR_SPECIALIST', 'ACCOUNTANT', 'NETWORK_MANAGER'],
  },
  {
    to: '/error-logs',
    label: 'Журнал ошибок',
    allowedRoles: ['HR_SPECIALIST', 'ACCOUNTANT', 'NETWORK_MANAGER', 'ADMINISTRATOR'],
  },
  {
    to: '/employees/new',
    label: 'Новый сотрудник',
    allowedRoles: ['HR_SPECIALIST', 'ADMINISTRATOR'],
  },
]

export function canAccessByRole(role: UserRole | undefined, allowedRoles?: UserRole[]) {
  if (!allowedRoles || allowedRoles.length === 0) {
    return true
  }

  if (!role) {
    return false
  }

  return allowedRoles.includes(role)
}

export function getVisibleNavigation(role: UserRole | undefined) {
  return protectedNavigation.filter((item) => canAccessByRole(role, item.allowedRoles))
}

export function canAccessRoute(role: UserRole | undefined, path: string) {
  const route = protectedNavigation.find((item) => item.to === path)
  if (!route) {
    return true
  }

  return canAccessByRole(role, route.allowedRoles)
}

export function getDefaultRouteForRole(role: UserRole | undefined) {
  switch (role) {
    case 'CLIENT':
    case 'BASIC_MASTER':
    case 'ADVANCED_MASTER':
      return '/bookings'
    case 'HR_SPECIALIST':
      return '/employees'
    case 'ACCOUNTANT':
      return '/payments'
    case 'NETWORK_MANAGER':
    case 'ADMINISTRATOR':
      return '/reports'
    default:
      return '/salons'
  }
}
