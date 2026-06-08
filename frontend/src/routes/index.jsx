import { lazy, Suspense } from 'react'
import { Routes, Route, Navigate } from 'react-router-dom'
import { ProtectedRoute } from './ProtectedRoute'
import { PublicRoute } from './PublicRoute'
import { ROUTES } from './routes'
import { LoadingSpinner } from '@components/common/LoadingSpinner'

// Pages - Lazy loaded for code splitting
const LandingPage = lazy(() =>
  import('@pages/LandingPage').then((m) => ({ default: m.LandingPage }))
)
const LoginPage = lazy(() => import('@pages/LoginPage').then((m) => ({ default: m.LoginPage })))
const RegisterPage = lazy(() =>
  import('@pages/RegisterPage').then((m) => ({ default: m.RegisterPage }))
)
const DashboardPage = lazy(() =>
  import('@pages/DashboardPage').then((m) => ({ default: m.DashboardPage }))
)
const ForgotPasswordPage = lazy(() =>
  import('@pages/ForgotPasswordPage').then((m) => ({ default: m.ForgotPasswordPage }))
)
const ContactsPage = lazy(() =>
  import('@pages/ContactsPage').then((m) => ({ default: m.ContactsPage }))
)
const DebtsPage = lazy(() => import('@pages/DebtsPage').then((m) => ({ default: m.DebtsPage })))
const SettingsPage = lazy(() =>
  import('@pages/SettingsPage').then((m) => ({ default: m.SettingsPage }))
)
const NotificationsPage = lazy(() =>
  import('@pages/NotificationsPage').then((m) => ({ default: m.NotificationsPage }))
)

// Placeholder component for pages not yet implemented
const PlaceholderPage = ({ title }) => (
  <div className="container mx-auto p-8">
    <h1 className="mb-6 text-3xl font-bold text-foreground">{title}</h1>
    <p className="text-muted-foreground">
      This page will be implemented in Phase 3 of the migration.
    </p>
  </div>
)

export const AppRoutes = () => {
  return (
    <Suspense fallback={<LoadingSpinner size="lg" message="Loading page..." className="py-12" />}>
      <Routes>
        {/* Public routes */}
        <Route path={ROUTES.HOME} element={<LandingPage />} />
        <Route
          path={ROUTES.LOGIN}
          element={
            <PublicRoute>
              <LoginPage />
            </PublicRoute>
          }
        />
        <Route
          path={ROUTES.REGISTER}
          element={
            <PublicRoute>
              <RegisterPage />
            </PublicRoute>
          }
        />
        <Route path={ROUTES.FORGOT_PASSWORD} element={<ForgotPasswordPage />} />
        <Route path={ROUTES.PRIVACY} element={<PlaceholderPage title="Privacy Policy" />} />
        <Route path={ROUTES.TERMS} element={<PlaceholderPage title="Terms of Service" />} />

        {/* Protected routes */}
        <Route
          path={ROUTES.DASHBOARD}
          element={
            <ProtectedRoute>
              <DashboardPage />
            </ProtectedRoute>
          }
        />
        <Route
          path={ROUTES.CONTACTS}
          element={
            <ProtectedRoute>
              <ContactsPage />
            </ProtectedRoute>
          }
        />
        <Route
          path={ROUTES.DEBTS}
          element={
            <ProtectedRoute>
              <DebtsPage />
            </ProtectedRoute>
          }
        />
        <Route
          path={ROUTES.NOTIFICATIONS}
          element={
            <ProtectedRoute>
              <NotificationsPage />
            </ProtectedRoute>
          }
        />
        <Route
          path={ROUTES.SETTINGS}
          element={
            <ProtectedRoute>
              <SettingsPage />
            </ProtectedRoute>
          }
        />

        {/* Catch all - redirect to home */}
        <Route path="*" element={<Navigate to={ROUTES.HOME} replace />} />
      </Routes>
    </Suspense>
  )
}
