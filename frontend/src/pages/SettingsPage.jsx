import { useState, useEffect } from 'react'
import { useSettingsStore } from '@stores/settingsStore'
import { useThemeStore } from '@stores/themeStore'
import { useAuthStore } from '@stores/authStore'
import { useNotificationsStore } from '@stores/notificationsStore'
import { useUserSettingsStore } from '@stores/userSettingsStore'
import { GeneralSettings } from '@components/settings/GeneralSettings'
import { NotificationSettings } from '@components/settings/NotificationSettings'

export const SettingsPage = () => {
  const user = useAuthStore((state) => state.user)
  const { theme, setTheme } = useThemeStore()
  const { language, setLanguage, resetSettings } = useSettingsStore()
  const { success, error: showError } = useNotificationsStore()

  // User settings from backend
  const {
    settings: userSettings,
    isLoading,
    isSaving,
    error: settingsError,
    errorType,
    validationErrors,
    fetchUserSettings,
    updateUserSettings,
    resetError,
  } = useUserSettingsStore()

  const [activeTab, setActiveTab] = useState('general')

  // Fetch settings on mount
  useEffect(() => {
    fetchUserSettings().catch((err) => {
      showError(err.message || 'Failed to load settings')
    })
  }, [fetchUserSettings, showError])

  const handleThemeChange = (newTheme) => {
    setTheme(newTheme)
    success('Theme updated successfully')
  }

  const handleLanguageChange = (newLanguage) => {
    setLanguage(newLanguage)
    success('Language updated successfully')
  }

  const handleGeneralSettingsSave = async (changes) => {
    try {
      await updateUserSettings(changes)
      success('General settings saved successfully')
    } catch (error) {
      showError(error.message || 'Failed to save general settings')
    }
  }

  const handleNotificationSettingsSave = async (changes) => {
    try {
      await updateUserSettings(changes)
      success('Notification settings saved successfully')
    } catch (error) {
      showError(error.message || 'Failed to save notification settings')
    }
  }

  const handleResetSettings = () => {
    if (window.confirm('Are you sure you want to reset all settings to defaults?')) {
      resetSettings()
      setTheme('light')
      success('Settings reset to defaults')
    }
  }

  return (
    <div className="mx-auto max-w-5xl space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-3xl font-bold text-foreground">Settings</h1>
        <p className="mt-1 text-muted-foreground">
          Manage your account preferences and application settings
        </p>
      </div>

      {/* Error Message */}
      {settingsError && (
        <div className="rounded-lg border border-destructive/50 bg-destructive/10 p-4">
          <div className="flex items-start justify-between">
            <div className="flex-1">
              <div className="flex items-center">
                <svg
                  className="mr-2 h-5 w-5 flex-shrink-0 text-destructive"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth="2"
                    d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                  />
                </svg>
                <div className="flex-1">
                  <span className="text-sm font-medium text-destructive">{settingsError}</span>
                  {errorType === 'network' && (
                    <p className="mt-1 text-xs text-destructive/80">
                      Please check your internet connection and try again.
                    </p>
                  )}
                  {errorType === 'server' && (
                    <p className="mt-1 text-xs text-destructive/80">
                      Our servers are experiencing issues. Please try again later or contact support
                      if the problem persists.
                    </p>
                  )}
                  {errorType === 'validation' && validationErrors && (
                    <div className="mt-2 space-y-1">
                      {Object.entries(validationErrors).map(([field, message]) => (
                        <p key={field} className="text-xs text-destructive/80">
                          <span className="font-medium">{field}:</span> {message}
                        </p>
                      ))}
                    </div>
                  )}
                </div>
              </div>
            </div>
            <div className="ml-4 flex items-start space-x-2">
              {errorType === 'network' && (
                <button
                  onClick={() => {
                    resetError()
                    fetchUserSettings().catch((err) => {
                      showError(err.message || 'Failed to load settings')
                    })
                  }}
                  className="btn-secondary text-xs"
                  disabled={isLoading}
                >
                  Retry
                </button>
              )}
              <button
                onClick={resetError}
                className="text-sm text-destructive hover:underline"
                aria-label="Dismiss error"
              >
                <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth="2"
                    d="M6 18L18 6M6 6l12 12"
                  />
                </svg>
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Tabs */}
      <div className="border-b border-border">
        <nav className="-mb-px flex space-x-8">
          <button
            onClick={() => setActiveTab('general')}
            className={`border-b-2 px-1 py-4 text-sm font-medium transition-colors ${
              activeTab === 'general'
                ? 'border-primary text-primary'
                : 'border-transparent text-muted-foreground hover:border-border hover:text-foreground'
            }`}
          >
            General
          </button>
          <button
            onClick={() => setActiveTab('notifications')}
            className={`border-b-2 px-1 py-4 text-sm font-medium transition-colors ${
              activeTab === 'notifications'
                ? 'border-primary text-primary'
                : 'border-transparent text-muted-foreground hover:border-border hover:text-foreground'
            }`}
          >
            Notifications
          </button>
          <button
            onClick={() => setActiveTab('account')}
            className={`border-b-2 px-1 py-4 text-sm font-medium transition-colors ${
              activeTab === 'account'
                ? 'border-primary text-primary'
                : 'border-transparent text-muted-foreground hover:border-border hover:text-foreground'
            }`}
          >
            Account
          </button>
        </nav>
      </div>

      {/* Tab Content */}
      <div className="space-y-6">
        {/* General Settings */}
        {activeTab === 'general' && (
          <GeneralSettings
            userSettings={userSettings}
            isLoading={isLoading}
            isSaving={isSaving}
            validationErrors={validationErrors}
            onSave={handleGeneralSettingsSave}
          />
        )}

        {/* Notification Settings */}
        {activeTab === 'notifications' && (
          <NotificationSettings
            userSettings={userSettings}
            isLoading={isLoading}
            isSaving={isSaving}
            validationErrors={validationErrors}
            onSave={handleNotificationSettingsSave}
            onRefreshSettings={() => fetchUserSettings()}
          />
        )}

        {/* Account Settings */}
        {activeTab === 'account' && (
          <div className="space-y-6">
            {/* User Information */}
            <div className="card p-6">
              <h2 className="mb-4 text-xl font-semibold text-foreground">Account Information</h2>

              <div className="space-y-4">
                <div className="flex items-center space-x-4">
                  <div className="flex h-16 w-16 items-center justify-center rounded-full bg-primary/10">
                    <span className="text-2xl font-medium text-primary">
                      {user?.first_name
                        ?.split(' ')
                        .map((n) => n[0])
                        .join('')
                        .toUpperCase() || 'U'}
                    </span>
                  </div>
                  <div>
                    <div className="text-lg font-medium text-foreground">
                      {user.first_name} {user.last_name}
                    </div>
                    <div className="text-sm text-muted-foreground">{user?.email || 'N/A'}</div>
                    <div className="text-sm text-muted-foreground">{user?.phone || 'N/A'}</div>
                  </div>
                </div>
              </div>
            </div>

            {/* Danger Zone */}
            <div className="card border-destructive/20 p-6">
              <h2 className="mb-4 text-xl font-semibold text-destructive">Danger Zone</h2>

              <div className="space-y-4">
                <div className="flex items-center justify-between rounded-lg border border-border p-4">
                  <div>
                    <div className="font-medium text-foreground">Reset Settings</div>
                    <div className="text-sm text-muted-foreground">
                      Reset all settings to their default values
                    </div>
                  </div>
                  <button onClick={handleResetSettings} className="btn-secondary">
                    Reset
                  </button>
                </div>

                <div className="flex items-center justify-between rounded-lg border border-destructive/50 bg-destructive/5 p-4">
                  <div>
                    <div className="font-medium text-foreground">Delete Account</div>
                    <div className="text-sm text-muted-foreground">
                      Permanently delete your account and all data
                    </div>
                  </div>
                  <button
                    onClick={() => alert('Account deletion functionality will be implemented soon')}
                    className="btn-destructive"
                  >
                    Delete Account
                  </button>
                </div>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
