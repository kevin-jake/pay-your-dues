import { useState, useEffect } from 'react'
import { useDeliveryNotificationsStore } from '@stores/deliveryNotificationsStore'
import { useNotificationsStore } from '@stores/notificationsStore'
import { LoadingSpinner } from '@components/common/LoadingSpinner'
import { NotificationScheduleRow } from './NotificationScheduleRow'
import { DebtNotificationSettingsForm } from './DebtNotificationSettingsForm'
import { SendNotificationModal } from './SendNotificationModal'
import { groupNotificationsBySlot } from '@utils/groupNotifications'

export const DebtNotificationsPanel = ({ debt }) => {
  const { success, error: showError } = useNotificationsStore()
  const {
    fetchByDebtList,
    getDebtNotifications,
    getDebtNotificationSettings,
    enableDebtNotifications,
    disableDebtNotifications,
    deleteNotificationSlot,
    enable,
    disable,
    sendManual,
    getManualSendLimits,
    isLoading,
    isActionLoading,
  } = useDeliveryNotificationsStore()

  const [notificationsEnabled, setNotificationsEnabled] = useState(true)
  const [showForm, setShowForm] = useState(false)
  const [initialFormSettings, setInitialFormSettings] = useState(null)
  const [isSavingForm, setIsSavingForm] = useState(false)
  const [showSendModal, setShowSendModal] = useState(false)
  const [sendLimits, setSendLimits] = useState(null)
  const [settingsLoading, setSettingsLoading] = useState(true)

  useEffect(() => {
    const load = async () => {
      setSettingsLoading(true)
      try {
        const [settings] = await Promise.all([
          getDebtNotificationSettings(debt.id),
          fetchByDebtList(debt.id),
        ])
        setNotificationsEnabled(settings.notifications_enabled)
        setInitialFormSettings(settings.settings)
      } catch (err) {
        showError(err.message || 'Failed to load notification settings')
      } finally {
        setSettingsLoading(false)
      }
    }
    load()
  }, [debt.id, fetchByDebtList, getDebtNotificationSettings, showError])

  const notifications = getDebtNotifications(debt.id)
  const groups = groupNotificationsBySlot(notifications)

  const loadLimits = async () => {
    try {
      const limits = await getManualSendLimits(debt.id)
      setSendLimits(limits)
    } catch {
      setSendLimits(null)
    }
  }

  const handleToggleMaster = async () => {
    if (notificationsEnabled) {
      if (!window.confirm('This will delete all scheduled reminders for this debt. Disable?')) return
      try {
        await disableDebtNotifications(debt.id)
        setNotificationsEnabled(false)
        success('Payment reminders disabled')
      } catch (err) {
        showError(err.message || 'Failed to disable notifications')
      }
    } else {
      // Show settings form
      setShowForm(true)
    }
  }

  const handleEnableWithSettings = async (settings) => {
    setIsSavingForm(true)
    try {
      await enableDebtNotifications(debt.id, settings)
      setNotificationsEnabled(true)
      setInitialFormSettings(settings)
      setShowForm(false)
      success('Payment reminders enabled')
    } catch (err) {
      showError(err.message || 'Failed to enable notifications')
    } finally {
      setIsSavingForm(false)
    }
  }

  const handleDeleteSlot = async (slotInfo) => {
    if (!window.confirm('Delete all channel notifications for this reminder slot?')) return
    try {
      await deleteNotificationSlot(debt.id, slotInfo)
      success('Reminder slot deleted')
    } catch (err) {
      showError(err.message || 'Failed to delete slot')
    }
  }

  const handleToggleChannel = async (notification) => {
    try {
      if (notification.enabled) {
        await disable(notification.id, debt.id)
        success('Channel notification disabled')
      } else {
        await enable(notification.id, debt.id)
        success('Channel notification enabled')
      }
    } catch (err) {
      showError(err.message || 'Failed to update channel')
    }
  }

  const handleSend = async ({ message, notificationType }) => {
    try {
      await sendManual({ debtListId: debt.id, message, notificationType })
      success('Notification sent')
      setShowSendModal(false)
      await loadLimits()
    } catch (err) {
      showError(err.message || 'Failed to send notification')
    }
  }

  const handleOpenSendModal = async () => {
    await loadLimits()
    setShowSendModal(true)
  }

  if (settingsLoading) {
    return (
      <div className="flex justify-center py-8">
        <LoadingSpinner message="Loading notifications…" />
      </div>
    )
  }

  return (
    <div className="space-y-4">
      {/* Master toggle header */}
      <div className="flex items-center justify-between rounded-lg border border-border p-4">
        <div>
          <p className="font-medium text-foreground">Payment Reminders</p>
          <p className="text-sm text-muted-foreground">
            {notificationsEnabled
              ? 'Reminders are scheduled based on due dates'
              : 'Reminders are disabled for this debt'}
          </p>
        </div>
        <div className="flex items-center gap-3">
          {notificationsEnabled && (
            <button
              onClick={handleOpenSendModal}
              disabled={isActionLoading}
              className="btn-secondary text-sm disabled:cursor-not-allowed disabled:opacity-50"
            >
              Send Now
            </button>
          )}
          <button
            onClick={handleToggleMaster}
            disabled={isActionLoading}
            className={`relative inline-flex h-7 w-12 items-center rounded-full transition-colors disabled:cursor-not-allowed disabled:opacity-50 ${
              notificationsEnabled ? 'bg-primary' : 'bg-muted'
            }`}
            aria-label={notificationsEnabled ? 'Disable reminders' : 'Enable reminders'}
          >
            <span
              className={`inline-block h-5 w-5 transform rounded-full bg-white transition-transform shadow ${
                notificationsEnabled ? 'translate-x-6' : 'translate-x-1'
              }`}
            />
          </button>
        </div>
      </div>

      {/* Settings form when enabling */}
      {!notificationsEnabled && showForm && (
        <DebtNotificationSettingsForm
          initialSettings={initialFormSettings}
          onSave={handleEnableWithSettings}
          onCancel={() => setShowForm(false)}
          isSaving={isSavingForm}
        />
      )}

      {/* Enable prompt when disabled and form not shown */}
      {!notificationsEnabled && !showForm && (
        <div className="rounded-lg border border-dashed border-border p-6 text-center">
          <p className="mb-3 text-sm text-muted-foreground">
            Enable reminders to receive notifications before payment due dates.
          </p>
          <button onClick={() => setShowForm(true)} className="btn-primary text-sm">
            Enable Reminders
          </button>
        </div>
      )}

      {/* Grouped slot table */}
      {notificationsEnabled && (
        <div className="rounded-lg border border-border">
          {isLoading ? (
            <div className="flex justify-center py-8">
              <LoadingSpinner message="Loading reminders…" />
            </div>
          ) : groups.length === 0 ? (
            <div className="p-6 text-center text-sm text-muted-foreground">
              No reminders scheduled yet. Reminders are auto-created when a debt is created.
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-border bg-muted/30 text-left text-muted-foreground">
                    <th className="w-8 pb-2 pl-2 pt-3" />
                    <th className="pb-2 pr-4 pt-3 font-medium">Installment</th>
                    <th className="pb-2 pr-4 pt-3 font-medium">Scheduled For</th>
                    <th className="pb-2 pr-4 pt-3 font-medium">Status</th>
                    <th className="pb-2 pr-4 pt-3 font-medium">Channels</th>
                    <th className="pb-2 pr-4 pt-3 font-medium">Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {groups.map((group) => (
                    <NotificationScheduleRow
                      key={group.key}
                      group={group}
                      onDelete={handleDeleteSlot}
                      onToggleChannel={handleToggleChannel}
                      isActionLoading={isActionLoading}
                    />
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      )}

      {showSendModal && (
        <SendNotificationModal
          debt={debt}
          onClose={() => setShowSendModal(false)}
          onSend={handleSend}
          isLoading={isActionLoading}
          sendLimits={sendLimits}
        />
      )}
    </div>
  )
}
