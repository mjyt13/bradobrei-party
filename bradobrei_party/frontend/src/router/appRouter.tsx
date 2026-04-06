import { createBrowserRouter, Navigate } from 'react-router-dom'
import { AppShell } from '../components/AppShell'
import { ProtectedRoute } from '../components/ProtectedRoute'
import { AuthPage } from '../pages/AuthPage'
import { BookingsPage } from '../pages/BookingsPage'
import { EmployeesPage } from '../pages/EmployeesPage'
import { EmployeeRegistrationPage } from '../pages/EmployeeRegistrationPage'
import { ErrorPage } from '../pages/ErrorPage'
import { ErrorLogsPage } from '../pages/ErrorLogsPage'
import { MaterialsPage } from '../pages/MaterialsPage'
import { PaymentsPage } from '../pages/PaymentsPage'
import { ReportsPage } from '../pages/ReportsPage'
import { SalonsPage } from '../pages/SalonsPage'
import { ServicesPage } from '../pages/ServicesPage'

export const appRouter = createBrowserRouter([
  {
    path: '/auth',
    element: <AuthPage />,
  },
  {
    element: <ProtectedRoute />,
    children: [
      {
        element: <AppShell />,
        children: [
          { index: true, element: <Navigate to="/reports" replace /> },
          { path: '/salons', element: <SalonsPage /> },
          { path: '/bookings', element: <BookingsPage /> },
          { path: '/services', element: <ServicesPage /> },
          { path: '/materials', element: <MaterialsPage /> },
          { path: '/payments', element: <PaymentsPage /> },
          { path: '/employees', element: <EmployeesPage /> },
          { path: '/reports', element: <ReportsPage /> },
          { path: '/error-logs', element: <ErrorLogsPage /> },
          { path: '/employees/new', element: <EmployeeRegistrationPage /> },
          { path: '*', element: <ErrorPage /> },
        ],
      },
    ],
  },
  {
    path: '*',
    element: <ErrorPage />,
  },
])
