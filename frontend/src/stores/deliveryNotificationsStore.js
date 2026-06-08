import { create } from 'zustand'
import { apiClient } from '@api/client'
import { transformNotification, transformNotifications } from '@utils/notificationTransform'

export const useDeliveryNotificationsStore = create((set, get) => ({
  notifications: [],
  debtNotifications: {},
  isLoading: false,
  isActionLoading: false,
  error: null,

  fetchAll: async (params = {}) => {
    try {
      set({ isLoading: true, error: null })
      const response = await apiClient.getNotifications(params)
      const notifications = transformNotifications(response)
      set({ notifications, isLoading: false })
      return notifications
    } catch (error) {
      set({ error: error.message, isLoading: false })
      throw error
    }
  },

  fetchByDebtList: async (debtListId) => {
    try {
      set({ isLoading: true, error: null })
      const response = await apiClient.getNotificationsByDebtList(debtListId)
      const notifications = transformNotifications(response)
      set((state) => ({
        debtNotifications: { ...state.debtNotifications, [debtListId]: notifications },
        isLoading: false,
      }))
      return notifications
    } catch (error) {
      set({ error: error.message, isLoading: false })
      throw error
    }
  },

  schedule: async (debtListId) => {
    try {
      set({ isActionLoading: true, error: null })
      await apiClient.scheduleNotifications(debtListId)
      set({ isActionLoading: false })
      await get().fetchByDebtList(debtListId)
    } catch (error) {
      set({ error: error.message, isActionLoading: false })
      throw error
    }
  },

  getDebtNotificationSettings: async (debtListId) => {
    return apiClient.getDebtNotificationSettings(debtListId)
  },

  enableDebtNotifications: async (debtListId, settings = null) => {
    try {
      set({ isActionLoading: true, error: null })
      await apiClient.enableDebtNotifications(debtListId, settings)
      set({ isActionLoading: false })
      await get().fetchByDebtList(debtListId)
    } catch (error) {
      set({ error: error.message, isActionLoading: false })
      throw error
    }
  },

  disableDebtNotifications: async (debtListId) => {
    try {
      set({ isActionLoading: true, error: null })
      await apiClient.disableDebtNotifications(debtListId)
      set((state) => ({
        debtNotifications: { ...state.debtNotifications, [debtListId]: [] },
        isActionLoading: false,
      }))
    } catch (error) {
      set({ error: error.message, isActionLoading: false })
      throw error
    }
  },

  deleteNotificationSlot: async (debtListId, slotInfo) => {
    try {
      set({ isActionLoading: true, error: null })
      await apiClient.deleteNotificationSlot(debtListId, slotInfo)
      set({ isActionLoading: false })
      await get().fetchByDebtList(debtListId)
    } catch (error) {
      set({ error: error.message, isActionLoading: false })
      throw error
    }
  },

  getManualSendLimits: async (debtListId) => {
    return apiClient.getManualSendLimits(debtListId)
  },

  sendManual: async ({ debtListId, message, notificationType }) => {
    try {
      set({ isActionLoading: true, error: null })
      await apiClient.sendManualNotification({ debtListId, message, notificationType })
      set({ isActionLoading: false })
      await get().fetchByDebtList(debtListId)
    } catch (error) {
      set({ error: error.message, isActionLoading: false })
      throw error
    }
  },

  enable: async (id, debtListId = null) => {
    try {
      set({ isActionLoading: true, error: null })
      await apiClient.enableNotification(id)
      set({ isActionLoading: false })
      await get().refresh(debtListId)
    } catch (error) {
      set({ error: error.message, isActionLoading: false })
      throw error
    }
  },

  disable: async (id, debtListId = null) => {
    try {
      set({ isActionLoading: true, error: null })
      await apiClient.disableNotification(id)
      set({ isActionLoading: false })
      await get().refresh(debtListId)
    } catch (error) {
      set({ error: error.message, isActionLoading: false })
      throw error
    }
  },

  delete: async (id, debtListId = null) => {
    try {
      set({ isActionLoading: true, error: null })
      await apiClient.deleteNotification(id)
      set({ isActionLoading: false })
      await get().refresh(debtListId)
    } catch (error) {
      set({ error: error.message, isActionLoading: false })
      throw error
    }
  },

  getNotification: async (id) => {
    const response = await apiClient.getNotification(id)
    return transformNotification(response)
  },

  refresh: async (debtListId = null) => {
    const promises = [get().fetchAll()]
    if (debtListId) {
      promises.push(get().fetchByDebtList(debtListId))
    }
    await Promise.all(promises)
  },

  getDebtNotifications: (debtListId) => {
    return get().debtNotifications[debtListId] || []
  },

  resetError: () => set({ error: null }),
}))
