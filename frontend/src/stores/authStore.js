import { create } from 'zustand'
import { apiClient, tokenManager } from '@api/client'
import { useUserSettingsStore } from './userSettingsStore'

export const useAuthStore = create((set, get) => ({
  // State
  user: null,
  isAuthenticated: false,
  isLoading: true,

  // Initialize auth state from localStorage and validate token
  init: async () => {
    const token = tokenManager.getToken()
    if (token) {
      try {
        // Validate token by making a request to the protected health endpoint
        await apiClient.healthCheck()

        // If we get here, the token is valid - retrieve user data
        const user = tokenManager.getUserData()
        set({
          isAuthenticated: true,
          isLoading: false,
          user,
        })

        // Fetch user settings if authenticated
        try {
          const userSettingsStore = useUserSettingsStore.getState()
          await userSettingsStore.fetchUserSettings()
        } catch (error) {
          // If settings fetch fails, it's not critical - user can still use the app
          // Backend will create default settings on first access
          console.warn('Failed to fetch user settings on app init:', error)
        }
      } catch (error) {
        // Token is invalid, remove it and user data
        console.log('Invalid token, removing from storage')
        tokenManager.removeToken()
        set({
          user: null,
          isAuthenticated: false,
          isLoading: false,
        })
      }
    } else {
      set({
        isLoading: false,
      })
    }
  },

  // Login action
  login: async (credentials) => {
    try {
      set({ isLoading: true })
      const response = await apiClient.login(credentials)

      // Store token and user data
      tokenManager.setToken(response.token)
      tokenManager.setUserData(response.user)

      // Update state
      set({
        user: response.user,
        isAuthenticated: true,
        isLoading: false,
      })

      // Fetch user settings after successful login
      try {
        const userSettingsStore = useUserSettingsStore.getState()
        await userSettingsStore.fetchUserSettings()
      } catch (error) {
        // If settings fetch fails, it's not critical - user can still use the app
        // Backend will create default settings on first access
        console.warn('Failed to fetch user settings after login:', error)
      }

      return response
    } catch (error) {
      set({ isLoading: false })
      throw error
    }
  },

  // Register action
  register: async (userData) => {
    try {
      set({ isLoading: true })
      const response = await apiClient.register(userData)

      set({ isLoading: false })
      return response
    } catch (error) {
      set({ isLoading: false })
      throw error
    }
  },

  // Set user data after successful login
  setUser: (user) => {
    tokenManager.setUserData(user)
    set({
      user,
      isAuthenticated: true,
      isLoading: false,
    })
  },

  // Logout action
  logout: () => {
    tokenManager.removeToken()
    set({
      user: null,
      isAuthenticated: false,
      isLoading: false,
    })
  },

  // Set loading state
  setLoading: (isLoading) => {
    set({ isLoading })
  },
}))
