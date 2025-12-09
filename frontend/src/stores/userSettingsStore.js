import { create } from 'zustand'
import { apiClient } from '@api/client'
import { useAuthStore } from '@stores/authStore'
import {
  transformBackendToFrontend,
  transformFrontendToBackend,
} from '@utils/userSettingsTransform'

// Get local timezone for default settings
const getLocalTimezone = () => {
  try {
    return Intl.DateTimeFormat().resolvedOptions().timeZone
  } catch (error) {
    // Fallback to UTC if timezone detection fails
    return 'UTC'
  }
}

// Default settings structure
const getDefaultSettings = () => ({
  id: null,
  userId: null,
  notificationEmail: true,
  notificationSms: false,
  notificationWebhook: false,
  notificationReminderDays: [7, 3, 1],
  notificationTime: '09:00:00',
  overdueReminderFrequency: 'daily',
  customEmailMessage: null,
  customSmsMessage: null,
  slackWebhookUrl: null,
  telegramChatId: null, // Set via Telegram bot subscription flow, not manually
  discordWebhookUrl: null,
  eventNotificationsEnabled: true,
  notifyContactOnPayment: true,
  notificationRecipient: 'both', // 'user', 'contact', or 'both'
  defaultCurrency: 'Php',
  timezone: getLocalTimezone(),
  createdAt: null,
  updatedAt: null,
})

// Error categorization helper
const categorizeError = (error) => {
  const errorMessage = error?.message || String(error || 'Unknown error')
  const errorString = errorMessage.toLowerCase()

  // Check error type
  if (errorString.includes('401') || errorString.includes('unauthorized')) {
    return { type: 'unauthorized', message: errorMessage }
  }
  if (errorString.includes('403') || errorString.includes('forbidden')) {
    return { type: 'forbidden', message: errorMessage }
  }
  if (errorString.includes('404') || errorString.includes('not found')) {
    return { type: 'not_found', message: errorMessage }
  }
  if (
    errorString.includes('422') ||
    errorString.includes('validation') ||
    errorString.includes('invalid')
  ) {
    return { type: 'validation', message: errorMessage, fields: error?.fields || null }
  }
  if (
    errorString.includes('500') ||
    errorString.includes('server error') ||
    errorString.includes('internal')
  ) {
    return { type: 'server', message: errorMessage }
  }
  if (
    errorString.includes('network') ||
    errorString.includes('fetch') ||
    errorString.includes('connection') ||
    errorString.includes('timeout')
  ) {
    return { type: 'network', message: errorMessage }
  }

  // Default to generic error
  return { type: 'generic', message: errorMessage }
}

export const useUserSettingsStore = create((set, get) => ({
  // State
  settings: getDefaultSettings(),
  isLoading: false,
  isSaving: false,
  error: null,
  errorType: null,
  validationErrors: null,
  lastFetched: null,

  // Fetch user settings from backend
  fetchUserSettings: async () => {
    try {
      set({ isLoading: true, error: null, errorType: null, validationErrors: null })
      const backendSettings = await apiClient.getUserSettings()
      const transformedSettings = transformBackendToFrontend(backendSettings)

      // If backend returns null or empty, use defaults (backend should create defaults, but handle gracefully)
      const finalSettings = transformedSettings || getDefaultSettings()

      set({
        settings: finalSettings,
        isLoading: false,
        lastFetched: new Date().toISOString(),
        error: null,
        errorType: null,
        validationErrors: null,
      })

      return finalSettings
    } catch (error) {
      const categorized = categorizeError(error)

      // Check if it's an authentication error
      if (categorized.type === 'unauthorized') {
        // Trigger logout flow
        const authStore = useAuthStore.getState()
        authStore.logout()
        set({
          error: categorized.message,
          errorType: categorized.type,
          isLoading: false,
        })
        throw error
      }

      // If settings don't exist (404), use defaults
      // Backend should create defaults on first access, but handle gracefully
      if (categorized.type === 'not_found') {
        console.warn(
          'User settings not found, using defaults. Backend will create on first update.'
        )
        const defaultSettings = getDefaultSettings()
        set({
          settings: defaultSettings,
          isLoading: false,
          error: null,
          errorType: null,
          validationErrors: null,
        })
        return defaultSettings
      }

      set({
        error: categorized.message,
        errorType: categorized.type,
        validationErrors: categorized.fields || null,
        isLoading: false,
      })
      throw error
    }
  },

  // Update user settings on backend
  updateUserSettings: async (updates) => {
    try {
      set({ isSaving: true, error: null, errorType: null, validationErrors: null })

      // Transform frontend format to backend format
      const backendUpdates = transformFrontendToBackend(updates)

      // Call API to update settings
      const backendResponse = await apiClient.updateUserSettings(backendUpdates)

      // Transform response back to frontend format
      const transformedSettings = transformBackendToFrontend(backendResponse)

      // Update local state with response
      set((state) => ({
        settings: transformedSettings || state.settings,
        isSaving: false,
        lastFetched: new Date().toISOString(),
        error: null,
        errorType: null,
        validationErrors: null,
      }))

      return transformedSettings
    } catch (error) {
      const categorized = categorizeError(error)

      // Check if it's an authentication error
      if (categorized.type === 'unauthorized') {
        // Trigger logout flow
        const authStore = useAuthStore.getState()
        authStore.logout()
      }

      set({
        error: categorized.message,
        errorType: categorized.type,
        validationErrors: categorized.fields || null,
        isSaving: false,
      })
      throw error
    }
  },

  // Sync settings with backend (fetches if stale or never fetched)
  syncSettings: async (force = false) => {
    const state = get()
    const now = new Date()
    const lastFetched = state.lastFetched ? new Date(state.lastFetched) : null

    // Check if we need to fetch (stale > 5 minutes or never fetched)
    const shouldFetch = force || !lastFetched || (now - lastFetched) / 1000 / 60 > 5 // 5 minutes in milliseconds

    if (shouldFetch) {
      return await state.fetchUserSettings()
    }

    return state.settings
  },

  // Reset error state
  resetError: () => {
    set({ error: null, errorType: null, validationErrors: null })
  },
}))
