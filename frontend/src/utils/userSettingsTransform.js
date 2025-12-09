// Get local timezone
const getLocalTimezone = () => {
  try {
    return Intl.DateTimeFormat().resolvedOptions().timeZone
  } catch (error) {
    // Fallback to UTC if timezone detection fails
    return 'UTC'
  }
}

/**
 * Transforms backend response (snake_case) to frontend format (camelCase)
 * @param {Object} backendSettings - Settings object from backend API
 * @returns {Object|null} Transformed settings object or null if input is invalid
 */
export const transformBackendToFrontend = (backendSettings) => {
  if (!backendSettings) return null

  return {
    id: backendSettings.id || null,
    userId: backendSettings.user_id || null,
    notificationEmail: backendSettings.notification_email ?? true,
    notificationSms: backendSettings.notification_sms ?? false,
    notificationWebhook: backendSettings.notification_webhook ?? false,
    notificationReminderDays: backendSettings.notification_reminder_days || [7, 3, 1],
    notificationTime: backendSettings.notification_time || '09:00:00',
    overdueReminderFrequency: backendSettings.overdue_reminder_frequency || 'daily',
    customEmailMessage: backendSettings.custom_email_message || null,
    customSmsMessage: backendSettings.custom_sms_message || null,
    slackWebhookUrl: backendSettings.slack_webhook_url || null,
    telegramChatId: backendSettings.telegram_chat_id || null, // Set via Telegram bot subscription
    discordWebhookUrl: backendSettings.discord_webhook_url || null,
    eventNotificationsEnabled: backendSettings.event_notifications_enabled ?? true,
    notifyContactOnPayment: backendSettings.notify_contact_on_payment ?? true,
    notificationRecipient: backendSettings.notification_recipient || 'both', // 'user', 'contact', or 'both'
    defaultCurrency: backendSettings.default_currency || 'Php',
    timezone: backendSettings.timezone || getLocalTimezone(),
    createdAt: backendSettings.created_at || null,
    updatedAt: backendSettings.updated_at || null,
  }
}

/**
 * Transforms frontend format (camelCase) to backend request format (snake_case)
 * @param {Object} frontendSettings - Settings object in frontend format
 * @returns {Object} Transformed settings object for backend API
 */
export const transformFrontendToBackend = (frontendSettings) => {
  if (!frontendSettings) return {}

  // Only include defined values (exclude null/undefined)
  const transformed = {}

  if (frontendSettings.notificationEmail !== undefined) {
    transformed.notification_email = frontendSettings.notificationEmail
  }
  if (frontendSettings.notificationSms !== undefined) {
    transformed.notification_sms = frontendSettings.notificationSms
  }
  if (frontendSettings.notificationWebhook !== undefined) {
    transformed.notification_webhook = frontendSettings.notificationWebhook
  }
  if (frontendSettings.notificationReminderDays !== undefined) {
    transformed.notification_reminder_days = frontendSettings.notificationReminderDays
  }
  if (frontendSettings.notificationTime !== undefined) {
    transformed.notification_time = frontendSettings.notificationTime
  }
  if (frontendSettings.overdueReminderFrequency !== undefined) {
    transformed.overdue_reminder_frequency = frontendSettings.overdueReminderFrequency
  }
  if (frontendSettings.customEmailMessage !== undefined) {
    transformed.custom_email_message = frontendSettings.customEmailMessage
  }
  if (frontendSettings.customSmsMessage !== undefined) {
    transformed.custom_sms_message = frontendSettings.customSmsMessage
  }
  if (frontendSettings.slackWebhookUrl !== undefined) {
    transformed.slack_webhook_url = frontendSettings.slackWebhookUrl
  }
  // Note: telegramChatId is managed via the Telegram bot subscription flow, not directly via API
  if (frontendSettings.discordWebhookUrl !== undefined) {
    transformed.discord_webhook_url = frontendSettings.discordWebhookUrl
  }
  if (frontendSettings.eventNotificationsEnabled !== undefined) {
    transformed.event_notifications_enabled = frontendSettings.eventNotificationsEnabled
  }
  if (frontendSettings.notifyContactOnPayment !== undefined) {
    transformed.notify_contact_on_payment = frontendSettings.notifyContactOnPayment
  }
  if (frontendSettings.notificationRecipient !== undefined) {
    transformed.notification_recipient = frontendSettings.notificationRecipient
  }
  if (frontendSettings.defaultCurrency !== undefined) {
    transformed.default_currency = frontendSettings.defaultCurrency
  }
  if (frontendSettings.timezone !== undefined) {
    transformed.timezone = frontendSettings.timezone
  }

  return transformed
}

/**
 * Maps backend notification fields to frontend notification preferences format
 * Used for compatibility with existing settingsStore notification structure
 * @param {Object} backendSettings - Settings object from backend (can be in either format)
 * @returns {Object} Notification preferences object
 */
export const mapNotificationPreferences = (backendSettings) => {
  if (!backendSettings) {
    return {
      email: true,
      push: false,
      reminders: true,
    }
  }

  // Handle both backend format (snake_case) and frontend format (camelCase)
  return {
    email: backendSettings.notificationEmail ?? backendSettings.notification_email ?? true,
    push: backendSettings.notificationWebhook ?? backendSettings.notification_webhook ?? false,
    reminders:
      backendSettings.eventNotificationsEnabled ??
      backendSettings.event_notifications_enabled ??
      true,
  }
}
