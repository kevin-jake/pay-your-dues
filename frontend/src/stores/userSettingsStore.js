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
  telegramBotToken: null,
  telegramChatId: null,
  discordWebhookUrl: null,
  eventNotificationsEnabled: true,
  notifyContactOnPayment: true,
  defaultCurrency: 'Php',
  timezone: getLocalTimezone(),
  createdAt: null,
  updatedAt: null,
})

export const useUserSettingsStore = create((set, get) => ({
  // State
  settings: getDefaultSettings(),
  isLoading: false,
  isSaving: false,
  error: null,
  lastFetched: null,

  // Fetch user settings from backend
  fetchUserSettings: async () => {
    try {
      set({ isLoading: true, error: null })
      const backendSettings = await apiClient.getUserSettings()
      const transformedSettings = transformBackendToFrontend(backendSettings)

      // If backend returns null or empty, use defaults (backend should create defaults, but handle gracefully)
      const finalSettings = transformedSettings || getDefaultSettings()

      set({
        settings: finalSettings,
        isLoading: false,
        lastFetched: new Date().toISOString(),
      })

      return finalSettings
    } catch (error) {
      const errorMessage = error.message || 'Failed to fetch user settings'

      // Check if it's an authentication error
      if (errorMessage.includes('401') || errorMessage.includes('Unauthorized')) {
        // Trigger logout flow
        const authStore = useAuthStore.getState()
        authStore.logout()
        set({
          error: errorMessage,
          isLoading: false,
        })
        throw error
      }

      // If settings don't exist (404) or other non-auth errors, use defaults
      // Backend should create defaults on first access, but handle gracefully
      if (errorMessage.includes('404') || errorMessage.includes('Not Found')) {
        console.warn(
          'User settings not found, using defaults. Backend will create on first update.'
        )
        const defaultSettings = getDefaultSettings()
        set({
          settings: defaultSettings,
          isLoading: false,
          error: null,
        })
        return defaultSettings
      }

      set({
        error: errorMessage,
        isLoading: false,
      })
      throw error
    }
  },

  // Update user settings on backend
  updateUserSettings: async (updates) => {
    try {
      set({ isSaving: true, error: null })

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
      }))

      return transformedSettings
    } catch (error) {
      const errorMessage = error.message || 'Failed to update user settings'

      // Check if it's an authentication error
      if (errorMessage.includes('401') || errorMessage.includes('Unauthorized')) {
        // Trigger logout flow
        const authStore = useAuthStore.getState()
        authStore.logout()
      }

      set({
        error: errorMessage,
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
    set({ error: null })
  },
}))
