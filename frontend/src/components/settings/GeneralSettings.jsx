import { useState, useEffect } from 'react'
import { useSettingsStore } from '@stores/settingsStore'
import { LoadingSpinner } from '@components/common/LoadingSpinner'

export const GeneralSettings = ({
  userSettings,
  isLoading,
  isSaving,
  validationErrors,
  onSave,
}) => {
  const { language, setLanguage } = useSettingsStore()

  // Local state for pending changes
  const [pendingChanges, setPendingChanges] = useState({
    defaultCurrency: userSettings?.defaultCurrency || 'Php',
    timezone: userSettings?.timezone || Intl.DateTimeFormat().resolvedOptions().timeZone,
  })

  // Update local state when userSettings change
  useEffect(() => {
    if (userSettings) {
      setPendingChanges({
        defaultCurrency: userSettings.defaultCurrency || 'Php',
        timezone: userSettings.timezone || Intl.DateTimeFormat().resolvedOptions().timeZone,
      })
    }
  }, [userSettings])

  const handleCurrencyChange = (newCurrency) => {
    const backendCurrency = newCurrency === 'PHP' ? 'Php' : newCurrency
    setPendingChanges((prev) => ({ ...prev, defaultCurrency: backendCurrency }))
  }

  const handleTimezoneChange = (newTimezone) => {
    setPendingChanges((prev) => ({ ...prev, timezone: newTimezone }))
  }

  const handleLanguageChange = (newLanguage) => {
    setLanguage(newLanguage)
  }

  const hasChanges = () => {
    if (!userSettings) return false
    return (
      pendingChanges.defaultCurrency !== userSettings.defaultCurrency ||
      pendingChanges.timezone !== userSettings.timezone
    )
  }

  const handleSave = async (e) => {
    e?.preventDefault()
    e?.stopPropagation()

    if (!hasChanges()) {
      return
    }

    if (onSave) {
      try {
        await onSave(pendingChanges)
      } catch (error) {
        // Error is handled by parent component
      }
    }
  }

  if (isLoading) {
    return (
      <div className="card p-6">
        <div className="flex items-center justify-center py-12">
          <LoadingSpinner size="lg" message="Loading settings..." />
        </div>
      </div>
    )
  }

  return (
    <div className="card p-6">
      <div className="mb-4 flex items-center justify-between">
        <h2 className="text-xl font-semibold text-foreground">General Settings</h2>
        {hasChanges() && (
          <button
            type="button"
            onClick={handleSave}
            disabled={isSaving}
            className="btn-primary disabled:cursor-not-allowed disabled:opacity-50"
          >
            {isSaving ? 'Saving...' : 'Save Changes'}
          </button>
        )}
      </div>

      {/* Currency */}
      <div className="mb-6">
        <label className="mb-2 block text-sm font-medium text-foreground">Currency</label>
        <select
          value={
            pendingChanges.defaultCurrency === 'Php'
              ? 'PHP'
              : pendingChanges.defaultCurrency || 'PHP'
          }
          onChange={(e) => handleCurrencyChange(e.target.value)}
          disabled={isSaving}
          className={`input max-w-xs disabled:cursor-not-allowed disabled:opacity-50 ${
            validationErrors?.default_currency ? 'border-destructive' : ''
          }`}
        >
          <option value="USD">USD - US Dollar ($)</option>
          <option value="EUR">EUR - Euro (€)</option>
          <option value="GBP">GBP - British Pound (£)</option>
          <option value="JPY">JPY - Japanese Yen (¥)</option>
          <option value="PHP">PHP - Philippine Peso (₱)</option>
          <option value="CAD">CAD - Canadian Dollar (C$)</option>
          <option value="AUD">AUD - Australian Dollar (A$)</option>
          <option value="INR">INR - Indian Rupee (₹)</option>
        </select>
        {validationErrors?.default_currency && (
          <p className="mt-1 text-sm text-destructive">{validationErrors.default_currency}</p>
        )}
        <p className="mt-1 text-sm text-muted-foreground">
          Select your preferred currency for displaying amounts
        </p>
      </div>

      {/* Language */}
      <div className="mb-6">
        <label className="mb-2 block text-sm font-medium text-foreground">Language</label>
        <select
          value={language}
          onChange={(e) => handleLanguageChange(e.target.value)}
          disabled={isSaving}
          className="input max-w-xs disabled:cursor-not-allowed disabled:opacity-50"
        >
          <option value="en">English</option>
          <option value="es">Español</option>
          <option value="fr">Français</option>
          <option value="de">Deutsch</option>
          <option value="ja">日本語</option>
          <option value="zh">中文</option>
        </select>
        <p className="mt-1 text-sm text-muted-foreground">Select your preferred language</p>
      </div>

      {/* Timezone */}
      <div>
        <label className="mb-2 block text-sm font-medium text-foreground">Timezone</label>
        <select
          value={pendingChanges.timezone}
          onChange={(e) => handleTimezoneChange(e.target.value)}
          disabled={isSaving}
          className={`input max-w-xs disabled:cursor-not-allowed disabled:opacity-50 ${
            validationErrors?.timezone ? 'border-destructive' : ''
          }`}
        >
          <option value="UTC">UTC</option>
          <option value="America/New_York">Eastern Time (ET)</option>
          <option value="America/Chicago">Central Time (CT)</option>
          <option value="America/Denver">Mountain Time (MT)</option>
          <option value="America/Los_Angeles">Pacific Time (PT)</option>
          <option value="Europe/London">London (GMT)</option>
          <option value="Europe/Paris">Paris (CET)</option>
          <option value="Asia/Tokyo">Tokyo (JST)</option>
          <option value="Asia/Shanghai">Shanghai (CST)</option>
          <option value="Asia/Dubai">Dubai (GST)</option>
          <option value="Australia/Sydney">Sydney (AEST)</option>
          <option value="America/Sao_Paulo">São Paulo (BRT)</option>
        </select>
        {validationErrors?.timezone && (
          <p className="mt-1 text-sm text-destructive">{validationErrors.timezone}</p>
        )}
        <p className="mt-1 text-sm text-muted-foreground">
          Select your timezone for notifications and reminders
        </p>
      </div>
    </div>
  )
}
