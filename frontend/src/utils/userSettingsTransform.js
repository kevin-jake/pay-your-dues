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
 * Transforms backend response (PascalCase) to frontend format (camelCase)
 * @param {Object} backendSettings - Settings object from backend API
 * @returns {Object|null} Transformed settings object or null if input is invalid
 */
export const transformBackendToFrontend = (backendSettings) => {
  if (!backendSettings) return null

  return {
    id: backendSettings.ID || backendSettings.id || null,
    userId: backendSettings.UserID || backendSettings.user_id || null,
    notificationEmail: backendSettings.NotificationEmail ?? true,
    notificationSms: backendSettings.NotificationSms ?? false,
    notificationWebhook: backendSettings.NotificationWebhook ?? false,
    notificationReminderDays: backendSettings.NotificationReminderDays || [7, 3, 1],
    notificationTime: backendSettings.NotificationTime || '09:00:00',
    overdueReminderFrequency: backendSettings.OverdueReminderFrequency || 'daily',
    customEmailMessage: backendSettings.CustomEmailMessage || null,
    customSmsMessage: backendSettings.CustomSmsMessage || null,
    slackWebhookUrl: backendSettings.SlackWebhookUrl || null,
    telegramBotToken: backendSettings.TelegramBotToken || null,
    telegramChatId: backendSettings.TelegramChatId || null,
    discordWebhookUrl: backendSettings.DiscordWebhookUrl || null,
    eventNotificationsEnabled: backendSettings.EventNotificationsEnabled ?? true,
    notifyContactOnPayment: backendSettings.NotifyContactOnPayment ?? true,
    defaultCurrency: backendSettings.DefaultCurrency || 'Php',
    timezone: backendSettings.Timezone || getLocalTimezone(),
    createdAt: backendSettings.CreatedAt || backendSettings.created_at || null,
    updatedAt: backendSettings.UpdatedAt || backendSettings.updated_at || null,
  }
}

/**
 * Transforms frontend format (camelCase) to backend request format (PascalCase)
 * @param {Object} frontendSettings - Settings object in frontend format
 * @returns {Object} Transformed settings object for backend API
 */
export const transformFrontendToBackend = (frontendSettings) => {
  if (!frontendSettings) return {}

  // Only include defined values (exclude null/undefined)
  const transformed = {}

  if (frontendSettings.notificationEmail !== undefined) {
    transformed.NotificationEmail = frontendSettings.notificationEmail
  }
  if (frontendSettings.notificationSms !== undefined) {
    transformed.NotificationSms = frontendSettings.notificationSms
  }
  if (frontendSettings.notificationWebhook !== undefined) {
    transformed.NotificationWebhook = frontendSettings.notificationWebhook
  }
  if (frontendSettings.notificationReminderDays !== undefined) {
    transformed.NotificationReminderDays = frontendSettings.notificationReminderDays
  }
  if (frontendSettings.notificationTime !== undefined) {
    transformed.NotificationTime = frontendSettings.notificationTime
  }
  if (frontendSettings.overdueReminderFrequency !== undefined) {
    transformed.OverdueReminderFrequency = frontendSettings.overdueReminderFrequency
  }
  if (frontendSettings.customEmailMessage !== undefined) {
    transformed.CustomEmailMessage = frontendSettings.customEmailMessage
  }
  if (frontendSettings.customSmsMessage !== undefined) {
    transformed.CustomSmsMessage = frontendSettings.customSmsMessage
  }
  if (frontendSettings.slackWebhookUrl !== undefined) {
    transformed.SlackWebhookUrl = frontendSettings.slackWebhookUrl
  }
  if (frontendSettings.telegramBotToken !== undefined) {
    transformed.TelegramBotToken = frontendSettings.telegramBotToken
  }
  if (frontendSettings.telegramChatId !== undefined) {
    transformed.TelegramChatId = frontendSettings.telegramChatId
  }
  if (frontendSettings.discordWebhookUrl !== undefined) {
    transformed.DiscordWebhookUrl = frontendSettings.discordWebhookUrl
  }
  if (frontendSettings.eventNotificationsEnabled !== undefined) {
    transformed.EventNotificationsEnabled = frontendSettings.eventNotificationsEnabled
  }
  if (frontendSettings.notifyContactOnPayment !== undefined) {
    transformed.NotifyContactOnPayment = frontendSettings.notifyContactOnPayment
  }
  if (frontendSettings.defaultCurrency !== undefined) {
    transformed.DefaultCurrency = frontendSettings.defaultCurrency
  }
  if (frontendSettings.timezone !== undefined) {
    transformed.Timezone = frontendSettings.timezone
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

  // Handle both backend format (PascalCase) and frontend format (camelCase)
  return {
    email: backendSettings.notificationEmail ?? backendSettings.NotificationEmail ?? true,
    push: backendSettings.notificationWebhook ?? backendSettings.NotificationWebhook ?? false,
    reminders:
      backendSettings.eventNotificationsEnabled ??
      backendSettings.EventNotificationsEnabled ??
      true,
  }
}

