import { useEffect } from 'react'
import { BrowserRouter } from 'react-router-dom'
import { AppRoutes } from './routes'
import { useAuthStore } from '@stores/authStore'
import { useUserSettingsStore } from '@stores/userSettingsStore'
import { useSettingsStore } from '@stores/settingsStore'
import { initializeTheme } from '@stores/themeStore'
import { Layout } from '@components/layout/Layout'

function App() {
  const initAuth = useAuthStore((state) => state.init)
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated)
  const userSettings = useUserSettingsStore((state) => state.settings)
  const setCurrency = useSettingsStore((state) => state.setCurrency)

  useEffect(() => {
    // Initialize theme from localStorage
    initializeTheme()

    // Initialize auth state
    initAuth()
  }, [initAuth])

  // Sync currency preference from backend settings
  useEffect(() => {
    if (isAuthenticated && userSettings?.defaultCurrency) {
      // Map backend currency format to frontend format (e.g., "Php" -> "PHP")
      const frontendCurrency =
        userSettings.defaultCurrency === 'Php' ? 'PHP' : userSettings.defaultCurrency
      setCurrency(frontendCurrency)
    }
  }, [isAuthenticated, userSettings?.defaultCurrency, setCurrency])

  return (
    <BrowserRouter>
      <Layout>
        <AppRoutes />
      </Layout>
    </BrowserRouter>
  )
}

export default App
