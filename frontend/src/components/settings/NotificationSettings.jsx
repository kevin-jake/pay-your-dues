import { useState, useEffect } from 'react'
import { mapNotificationPreferences } from '@utils/userSettingsTransform'
import { LoadingSpinner } from '@components/common/LoadingSpinner'
import { NotificationToggle } from './NotificationToggle'
import { WebhookConfiguration } from './WebhookConfiguration'
import { NotificationSchedule } from './NotificationSchedule'

export const NotificationSettings = ({ userSettings, isLoading, isSaving, onSave }) => {
  // Map backend settings to frontend notification format for UI
  const notifications = mapNotificationPreferences(userSettings)
  const [localNotifications, setLocalNotifications] = useState(notifications)

  // Local state for pending changes
  const [pendingChanges, setPendingChanges] = useState({
    notificationEmail: userSettings?.notificationEmail ?? true,
    notificationSms: userSettings?.notificationSms ?? false,
    notificationWebhook: userSettings?.notificationWebhook ?? false,
    eventNotificationsEnabled: userSettings?.eventNotificationsEnabled ?? true,
    notifyContactOnPayment: userSettings?.notifyContactOnPayment ?? true,
    slackWebhookUrl: userSettings?.slackWebhookUrl || '',
    telegramBotToken: userSettings?.telegramBotToken || '',
    telegramChatId: userSettings?.telegramChatId || '',
    discordWebhookUrl: userSettings?.discordWebhookUrl || '',
    notificationReminderDays: userSettings?.notificationReminderDays || [7, 3, 1],
    notificationTime: userSettings?.notificationTime || '09:00:00',
    overdueReminderFrequency: userSettings?.overdueReminderFrequency || 'daily',
    customEmailMessage: userSettings?.customEmailMessage || '',
    customSmsMessage: userSettings?.customSmsMessage || '',
  })

  // Update local state when userSettings change
  useEffect(() => {
    const mapped = mapNotificationPreferences(userSettings)
    setLocalNotifications(mapped)

    if (userSettings) {
      setPendingChanges({
        notificationEmail: userSettings.notificationEmail ?? true,
        notificationSms: userSettings.notificationSms ?? false,
        notificationWebhook: userSettings.notificationWebhook ?? false,
        eventNotificationsEnabled: userSettings.eventNotificationsEnabled ?? true,
        notifyContactOnPayment: userSettings.notifyContactOnPayment ?? true,
        slackWebhookUrl: userSettings.slackWebhookUrl || '',
        telegramBotToken: userSettings.telegramBotToken || '',
        telegramChatId: userSettings.telegramChatId || '',
        discordWebhookUrl: userSettings.discordWebhookUrl || '',
        notificationReminderDays: userSettings.notificationReminderDays || [7, 3, 1],
        notificationTime: userSettings.notificationTime || '09:00:00',
        overdueReminderFrequency: userSettings.overdueReminderFrequency || 'daily',
        customEmailMessage: userSettings.customEmailMessage || '',
        customSmsMessage: userSettings.customSmsMessage || '',
      })
    }
  }, [userSettings])

  const handleNotificationChange = (key, value) => {
    const updated = { ...localNotifications, [key]: value }
    setLocalNotifications(updated)

    // Map to backend fields
    if (key === 'email') {
      setPendingChanges((prev) => ({ ...prev, notificationEmail: value }))
    } else if (key === 'push') {
      setPendingChanges((prev) => ({ ...prev, notificationWebhook: value }))
    } else if (key === 'reminders') {
      setPendingChanges((prev) => ({ ...prev, eventNotificationsEnabled: value }))
    }
  }

  const handleSmsToggle = (value) => {
    setPendingChanges((prev) => ({ ...prev, notificationSms: value }))
  }

  const handleEventNotificationToggle = (value) => {
    setPendingChanges((prev) => ({ ...prev, eventNotificationsEnabled: value }))
  }

  const handleNotifyContactToggle = (value) => {
    setPendingChanges((prev) => ({ ...prev, notifyContactOnPayment: value }))
  }

  const handleWebhookInputChange = (field, value) => {
    setPendingChanges((prev) => ({ ...prev, [field]: value || null }))
  }

  const handleAdvancedInputChange = (field, value) => {
    setPendingChanges((prev) => ({ ...prev, [field]: value }))
  }

  const hasChanges = () => {
    if (!userSettings) return false
    return (
      pendingChanges.notificationEmail !== (userSettings.notificationEmail ?? true) ||
      pendingChanges.notificationSms !== (userSettings.notificationSms ?? false) ||
      pendingChanges.notificationWebhook !== (userSettings.notificationWebhook ?? false) ||
      pendingChanges.eventNotificationsEnabled !==
        (userSettings.eventNotificationsEnabled ?? true) ||
      pendingChanges.notifyContactOnPayment !== (userSettings.notifyContactOnPayment ?? true) ||
      pendingChanges.slackWebhookUrl !== (userSettings.slackWebhookUrl || '') ||
      pendingChanges.telegramBotToken !== (userSettings.telegramBotToken || '') ||
      pendingChanges.telegramChatId !== (userSettings.telegramChatId || '') ||
      pendingChanges.discordWebhookUrl !== (userSettings.discordWebhookUrl || '') ||
      JSON.stringify(pendingChanges.notificationReminderDays) !==
        JSON.stringify(userSettings.notificationReminderDays || [7, 3, 1]) ||
      pendingChanges.notificationTime !== (userSettings.notificationTime || '09:00:00') ||
      pendingChanges.overdueReminderFrequency !==
        (userSettings.overdueReminderFrequency || 'daily') ||
      pendingChanges.customEmailMessage !== (userSettings.customEmailMessage || '') ||
      pendingChanges.customSmsMessage !== (userSettings.customSmsMessage || '')
    )
  }

  const handleSave = async (e) => {
    e?.preventDefault()
    e?.stopPropagation()

    console.log('Save button clicked')
    console.log('hasChanges():', hasChanges())
    console.log('pendingChanges:', pendingChanges)
    console.log('userSettings:', userSettings)

    if (!hasChanges()) {
      console.warn('No changes detected, save aborted')
      return
    }

    console.log('Saving notification settings:', pendingChanges)
    if (onSave) {
      try {
        await onSave(pendingChanges)
        console.log('Save completed successfully')
      } catch (error) {
        console.error('Save failed:', error)
      }
    } else {
      console.error('onSave handler is not provided')
    }
  }

  if (isLoading) {
    return (
      <div className="card p-6">
        <div className="flex items-center justify-center py-12">
          <LoadingSpinner size="lg" message="Loading notification settings..." />
        </div>
      </div>
    )
  }

  const webhookInputs = {
    slackWebhookUrl: pendingChanges.slackWebhookUrl,
    telegramBotToken: pendingChanges.telegramBotToken,
    telegramChatId: pendingChanges.telegramChatId,
    discordWebhookUrl: pendingChanges.discordWebhookUrl,
  }

  const advancedInputs = {
    notificationReminderDays: pendingChanges.notificationReminderDays,
    notificationTime: pendingChanges.notificationTime,
    overdueReminderFrequency: pendingChanges.overdueReminderFrequency,
  }

  return (
    <div className="card p-6">
      <div className="mb-4 flex items-center justify-between">
        <h2 className="text-xl font-semibold text-foreground">Notification Preferences</h2>
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

      {isSaving && (
        <div className="mb-4 flex items-center rounded-lg bg-primary/10 p-3 text-sm text-primary">
          <svg className="mr-2 h-4 w-4 animate-spin" fill="none" viewBox="0 0 24 24">
            <circle
              className="opacity-25"
              cx="12"
              cy="12"
              r="10"
              stroke="currentColor"
              strokeWidth="4"
            />
            <path
              className="opacity-75"
              fill="currentColor"
              d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
            />
          </svg>
          Saving changes...
        </div>
      )}

      <div className="space-y-4">
        {/* Email Notifications */}
        <div className="rounded-lg border border-border">
          <div className="p-4">
            <NotificationToggle
              label="Email Notifications"
              description="Receive notifications via email"
              enabled={localNotifications.email}
              onChange={(value) => handleNotificationChange('email', value)}
              disabled={isSaving}
            />
          </div>
          {localNotifications.email && (
            <div className="border-t border-border p-4">
              <label className="mb-2 block text-sm font-medium text-foreground">
                Custom Email Message Template
              </label>
              <textarea
                value={pendingChanges.customEmailMessage}
                onChange={(e) => handleAdvancedInputChange('customEmailMessage', e.target.value)}
                placeholder="Enter custom email message template..."
                rows={4}
                disabled={isSaving}
                className="input disabled:cursor-not-allowed disabled:opacity-50"
              />
              <p className="mt-1 text-xs text-muted-foreground">
                Custom message template for email notifications (optional)
              </p>
            </div>
          )}
        </div>

        {/* Push Notifications */}
        <NotificationToggle
          label="Push Notifications"
          description="Receive push notifications in your browser"
          enabled={localNotifications.push}
          onChange={(value) => handleNotificationChange('push', value)}
          disabled={isSaving}
        />

        {/* SMS Notifications */}
        <div className="rounded-lg border border-border">
          <div className="p-4">
            <NotificationToggle
              label="SMS Notifications"
              description="Receive notifications via SMS (requires SMS provider configuration)"
              enabled={pendingChanges.notificationSms}
              onChange={handleSmsToggle}
              disabled={isSaving}
            />
          </div>
          {pendingChanges.notificationSms && (
            <div className="border-t border-border p-4">
              <label className="mb-2 block text-sm font-medium text-foreground">
                Custom SMS Message Template
              </label>
              <textarea
                value={pendingChanges.customSmsMessage}
                onChange={(e) => handleAdvancedInputChange('customSmsMessage', e.target.value)}
                placeholder="Enter custom SMS message template..."
                rows={4}
                disabled={isSaving}
                className="input disabled:cursor-not-allowed disabled:opacity-50"
              />
              <p className="mt-1 text-xs text-muted-foreground">
                Custom message template for SMS notifications (optional)
              </p>
            </div>
          )}
        </div>

        {/* Payment Reminders */}
        <NotificationToggle
          label="Payment Reminders"
          description="Get reminders for upcoming due dates"
          enabled={localNotifications.reminders}
          onChange={(value) => handleNotificationChange('reminders', value)}
          disabled={isSaving}
        />

        {/* Event Notifications */}
        <div className="mt-6 border-t border-border pt-6">
          <h3 className="mb-4 text-lg font-semibold text-foreground">Event Notifications</h3>

          <div className="space-y-4">
            <NotificationToggle
              label="Enable Event Notifications"
              description="Receive notifications for payment events and updates"
              enabled={pendingChanges.eventNotificationsEnabled}
              onChange={handleEventNotificationToggle}
              disabled={isSaving}
            />

            <NotificationToggle
              label="Notify Contact on Payment"
              description="Send notification to contact when payment is received"
              enabled={pendingChanges.notifyContactOnPayment}
              onChange={handleNotifyContactToggle}
              disabled={isSaving}
            />
          </div>
        </div>

        {/* Webhook Configuration */}
        <WebhookConfiguration
          webhookInputs={webhookInputs}
          onInputChange={handleWebhookInputChange}
          isSaving={isSaving}
        />

        {/* Notification Schedule */}
        <NotificationSchedule
          advancedInputs={advancedInputs}
          onInputChange={handleAdvancedInputChange}
          isSaving={isSaving}
        />
      </div>
    </div>
  )
}
