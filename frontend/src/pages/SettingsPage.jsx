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
    console.log('handleGeneralSettingsSave called with:', changes)
    try {
      console.log('Calling updateUserSettings...')
      const result = await updateUserSettings(changes)
      console.log('updateUserSettings result:', result)
      success('General settings saved successfully')
    } catch (error) {
      console.error('Error in handleGeneralSettingsSave:', error)
      showError(error.message || 'Failed to save general settings')
    }
  }

  const handleNotificationSettingsSave = async (changes) => {
    console.log('handleNotificationSettingsSave called with:', changes)
    try {
      console.log('Calling updateUserSettings...')
      const result = await updateUserSettings(changes)
      console.log('updateUserSettings result:', result)
      success('Notification settings saved successfully')
    } catch (error) {
      console.error('Error in handleNotificationSettingsSave:', error)
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
          <div className="flex items-center justify-between">
            <div className="flex items-center">
              <svg
                className="mr-2 h-5 w-5 text-destructive"
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
              <span className="text-sm text-destructive">{settingsError}</span>
            </div>
            <button onClick={resetError} className="text-sm text-destructive hover:underline">
              Dismiss
            </button>
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
            onClick={() => setActiveTab('appearance')}
            className={`border-b-2 px-1 py-4 text-sm font-medium transition-colors ${
              activeTab === 'appearance'
                ? 'border-primary text-primary'
                : 'border-transparent text-muted-foreground hover:border-border hover:text-foreground'
            }`}
          >
            Appearance
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
            onSave={handleGeneralSettingsSave}
          />
        )}

        {/* Appearance Settings */}
        {activeTab === 'appearance' && (
          <div className="space-y-6">
            <div className="card p-6">
              <h2 className="mb-4 text-xl font-semibold text-foreground">Appearance</h2>

              <div>
                <label className="mb-3 block text-sm font-medium text-foreground">Theme</label>
                <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
                  {/* Light Theme */}
                  <button
                    onClick={() => handleThemeChange('light')}
                    className={`rounded-lg border-2 p-4 text-left transition-colors ${
                      theme === 'light'
                        ? 'border-primary bg-primary/5'
                        : 'border-border hover:border-primary/50'
                    }`}
                  >
                    <div className="mb-2 flex items-center justify-center rounded-lg bg-white p-4">
                      <svg
                        className="h-8 w-8 text-yellow-500"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth="2"
                          d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z"
                        />
                      </svg>
                    </div>
                    <div className="text-center">
                      <div className="font-medium text-foreground">Light</div>
                      <div className="text-sm text-muted-foreground">Bright and clean</div>
                    </div>
                  </button>

                  {/* Dark Theme */}
                  <button
                    onClick={() => handleThemeChange('dark')}
                    className={`rounded-lg border-2 p-4 text-left transition-colors ${
                      theme === 'dark'
                        ? 'border-primary bg-primary/5'
                        : 'border-border hover:border-primary/50'
                    }`}
                  >
                    <div className="mb-2 flex items-center justify-center rounded-lg bg-gray-900 p-4">
                      <svg
                        className="h-8 w-8 text-blue-400"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth="2"
                          d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z"
                        />
                      </svg>
                    </div>
                    <div className="text-center">
                      <div className="font-medium text-foreground">Dark</div>
                      <div className="text-sm text-muted-foreground">Easy on the eyes</div>
                    </div>
                  </button>

                  {/* System Theme */}
                  <button
                    onClick={() => handleThemeChange('system')}
                    className={`rounded-lg border-2 p-4 text-left transition-colors ${
                      theme === 'system'
                        ? 'border-primary bg-primary/5'
                        : 'border-border hover:border-primary/50'
                    }`}
                  >
                    <div className="mb-2 flex items-center justify-center rounded-lg bg-gradient-to-r from-white to-gray-900 p-4">
                      <svg
                        className="h-8 w-8 text-gray-600"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth="2"
                          d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"
                        />
                      </svg>
                    </div>
                    <div className="text-center">
                      <div className="font-medium text-foreground">System</div>
                      <div className="text-sm text-muted-foreground">Match device</div>
                    </div>
                  </button>
                </div>
                <p className="mt-3 text-sm text-muted-foreground">
                  Choose your preferred color scheme
                </p>
              </div>
            </div>
          </div>
        )}

        {/* Notification Settings */}
        {activeTab === 'notifications' && (
          <NotificationSettings
            userSettings={userSettings}
            isLoading={isLoading}
            isSaving={isSaving}
            onSave={handleNotificationSettingsSave}
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
                      {user?.username
                        ?.split(' ')
                        .map((n) => n[0])
                        .join('')
                        .toUpperCase() || 'U'}
                    </span>
                  </div>
                  <div>
                    <div className="text-lg font-medium text-foreground">
                      {user?.username || 'User'}
                    </div>
                    <div className="text-sm text-muted-foreground">{user?.email || 'N/A'}</div>
                  </div>
                </div>

                <div className="rounded-lg border border-border p-4">
                  <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                    <div>
                      <div className="text-sm font-medium text-muted-foreground">Username</div>
                      <div className="mt-1 text-foreground">{user?.username || 'N/A'}</div>
                    </div>
                    <div>
                      <div className="text-sm font-medium text-muted-foreground">Email</div>
                      <div className="mt-1 text-foreground">{user?.email || 'N/A'}</div>
                    </div>
                    <div>
                      <div className="text-sm font-medium text-muted-foreground">
                        Account Created
                      </div>
                      <div className="mt-1 text-foreground">
                        {user?.created_at ? new Date(user.created_at).toLocaleDateString() : 'N/A'}
                      </div>
                    </div>
                    <div>
                      <div className="text-sm font-medium text-muted-foreground">Last Updated</div>
                      <div className="mt-1 text-foreground">
                        {user?.updated_at ? new Date(user.updated_at).toLocaleDateString() : 'N/A'}
                      </div>
                    </div>
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
