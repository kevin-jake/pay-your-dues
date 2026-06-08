import { useState, useEffect, useMemo } from 'react'
import { useDeliveryNotificationsStore } from '@stores/deliveryNotificationsStore'
import { useDebtsStore } from '@stores/debtsStore'
import { useNotificationsStore } from '@stores/notificationsStore'
import { LoadingSpinner } from '@components/common/LoadingSpinner'
import { EmptyState } from '@components/common/EmptyState'
import {
  NotificationStatusBadge,
  NotificationTypeBadge,
} from '@components/notifications/NotificationStatusBadge'
import { DebtDetailsModal } from '@components/debts/DebtDetailsModal'
import { formatDate } from '@utils/formatters'

const STATUS_OPTIONS = [
  { value: '', label: 'All statuses' },
  { value: 'pending', label: 'Pending' },
  { value: 'queued', label: 'Queued' },
  { value: 'sent', label: 'Sent' },
  { value: 'failed', label: 'Failed' },
  { value: 'skipped', label: 'Skipped' },
  { value: 'cancelled', label: 'Cancelled' },
]

const TYPE_OPTIONS = [
  { value: '', label: 'All types' },
  { value: 'email', label: 'Email' },
  { value: 'sms', label: 'SMS' },
  { value: 'slack', label: 'Slack' },
  { value: 'telegram', label: 'Telegram' },
  { value: 'discord', label: 'Discord' },
]

const ITEMS_PER_PAGE = 10

export const NotificationsPage = () => {
  const { success, error: showError } = useNotificationsStore()
  const { debts, fetchDebts } = useDebtsStore()
  const {
    notifications,
    fetchAll,
    enable,
    disable,
    isLoading,
    isActionLoading,
  } = useDeliveryNotificationsStore()

  const [statusFilter, setStatusFilter] = useState('')
  const [typeFilter, setTypeFilter] = useState('')
  const [debtFilter, setDebtFilter] = useState('')
  const [currentPage, setCurrentPage] = useState(1)
  const [selectedNotification, setSelectedNotification] = useState(null)
  const [selectedDebt, setSelectedDebt] = useState(null)

  useEffect(() => {
    fetchDebts().catch(() => {})
  }, [fetchDebts])

  useEffect(() => {
    const params = { limit: 100 }
    if (statusFilter) params.status = statusFilter
    if (debtFilter) params.debtListId = debtFilter

    fetchAll(params).catch((err) => {
      showError(err.message || 'Failed to load notifications')
    })
    setCurrentPage(1)
  }, [statusFilter, debtFilter, fetchAll, showError])

  const debtMap = useMemo(() => {
    const map = {}
    debts.forEach((debt) => {
      map[debt.id] = debt
    })
    return map
  }, [debts])

  const filteredNotifications = useMemo(() => {
    let result = [...notifications]
    if (typeFilter) {
      result = result.filter((n) => n.notificationType === typeFilter)
    }
    return result.sort((a, b) => {
      const dateA = new Date(a.scheduledFor || a.createdAt)
      const dateB = new Date(b.scheduledFor || b.createdAt)
      return dateB - dateA
    })
  }, [notifications, typeFilter])

  const totalPages = Math.ceil(filteredNotifications.length / ITEMS_PER_PAGE)
  const paginatedNotifications = filteredNotifications.slice(
    (currentPage - 1) * ITEMS_PER_PAGE,
    currentPage * ITEMS_PER_PAGE
  )

  const handleToggleEnabled = async (notification) => {
    try {
      if (notification.enabled) {
        await disable(notification.id, notification.debtListId)
        success('Notification disabled')
      } else {
        await enable(notification.id, notification.debtListId)
        success('Notification enabled')
      }
    } catch (err) {
      showError(err.message || 'Failed to update notification')
    }
  }

  const handleOpenDebt = (debtListId) => {
    const debt = debtMap[debtListId]
    if (debt) {
      setSelectedDebt(debt)
    }
  }

  return (
    <div className="mx-auto max-w-7xl space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-foreground">Notifications</h1>
        <p className="mt-1 text-muted-foreground">
          View and manage payment reminder notifications across all debts
        </p>
      </div>

      {/* Filters */}
      <div className="card p-4">
        <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
          <div>
            <label className="mb-1 block text-sm font-medium text-foreground">Status</label>
            <select
              value={statusFilter}
              onChange={(e) => setStatusFilter(e.target.value)}
              className="input"
            >
              {STATUS_OPTIONS.map((opt) => (
                <option key={opt.value} value={opt.value}>
                  {opt.label}
                </option>
              ))}
            </select>
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium text-foreground">Type</label>
            <select
              value={typeFilter}
              onChange={(e) => setTypeFilter(e.target.value)}
              className="input"
            >
              {TYPE_OPTIONS.map((opt) => (
                <option key={opt.value} value={opt.value}>
                  {opt.label}
                </option>
              ))}
            </select>
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium text-foreground">Debt</label>
            <select
              value={debtFilter}
              onChange={(e) => setDebtFilter(e.target.value)}
              className="input"
            >
              <option value="">All debts</option>
              {debts.map((debt) => (
                <option key={debt.id} value={debt.id}>
                  {debt.contact?.name || debt.description || debt.id}
                </option>
              ))}
            </select>
          </div>
        </div>
      </div>

      {/* Table */}
      <div className="card overflow-hidden">
        {isLoading ? (
          <div className="flex justify-center py-12">
            <LoadingSpinner size="lg" message="Loading notifications..." />
          </div>
        ) : paginatedNotifications.length === 0 ? (
          <EmptyState
            title="No notifications found"
            description="Notifications will appear here when reminders are scheduled for your debts."
          />
        ) : (
          <>
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead className="border-b border-border bg-muted/30">
                  <tr className="text-left text-muted-foreground">
                    <th className="px-4 py-3 font-medium">Date</th>
                    <th className="px-4 py-3 font-medium">Debt / Contact</th>
                    <th className="px-4 py-3 font-medium">Type</th>
                    <th className="px-4 py-3 font-medium">Schedule</th>
                    <th className="px-4 py-3 font-medium">Installment</th>
                    <th className="px-4 py-3 font-medium">Status</th>
                    <th className="px-4 py-3 font-medium">Enabled</th>
                    <th className="px-4 py-3 font-medium">Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {paginatedNotifications.map((notification) => {
                    const debt = debtMap[notification.debtListId]
                    return (
                      <tr key={notification.id} className="border-b border-border/50">
                        <td className="px-4 py-3 text-foreground">
                          {notification.scheduledFor
                            ? formatDate(notification.scheduledFor)
                            : formatDate(notification.createdAt)}
                        </td>
                        <td className="px-4 py-3">
                          {debt ? (
                            <button
                              onClick={() => handleOpenDebt(notification.debtListId)}
                              className="text-left text-primary hover:underline"
                            >
                              {debt.contact?.name || debt.description || 'View debt'}
                            </button>
                          ) : (
                            <span className="text-muted-foreground">Unknown debt</span>
                          )}
                        </td>
                        <td className="px-4 py-3">
                          <NotificationTypeBadge type={notification.notificationType} />
                        </td>
                        <td className="px-4 py-3 capitalize text-foreground">
                          {notification.scheduleType || '—'}
                        </td>
                        <td className="px-4 py-3 text-foreground">
                          {notification.installmentNumber != null
                            ? `#${notification.installmentNumber}`
                            : '—'}
                        </td>
                        <td className="px-4 py-3">
                          <NotificationStatusBadge status={notification.status} />
                        </td>
                        <td className="px-4 py-3">
                          <button
                            onClick={() => handleToggleEnabled(notification)}
                            disabled={isActionLoading}
                            className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors disabled:cursor-not-allowed disabled:opacity-50 ${
                              notification.enabled ? 'bg-primary' : 'bg-muted'
                            }`}
                          >
                            <span
                              className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${
                                notification.enabled ? 'translate-x-6' : 'translate-x-1'
                              }`}
                            />
                          </button>
                        </td>
                        <td className="px-4 py-3">
                          <button
                            onClick={() => setSelectedNotification(notification)}
                            className="text-xs text-primary hover:underline"
                          >
                            Details
                          </button>
                        </td>
                      </tr>
                    )
                  })}
                </tbody>
              </table>
            </div>

            {totalPages > 1 && (
              <div className="flex items-center justify-between border-t border-border px-4 py-3">
                <span className="text-sm text-muted-foreground">
                  Page {currentPage} of {totalPages} ({filteredNotifications.length} total)
                </span>
                <div className="flex gap-2">
                  <button
                    onClick={() => setCurrentPage((p) => Math.max(1, p - 1))}
                    disabled={currentPage === 1}
                    className="btn-secondary text-sm disabled:cursor-not-allowed disabled:opacity-50"
                  >
                    Previous
                  </button>
                  <button
                    onClick={() => setCurrentPage((p) => Math.min(totalPages, p + 1))}
                    disabled={currentPage === totalPages}
                    className="btn-secondary text-sm disabled:cursor-not-allowed disabled:opacity-50"
                  >
                    Next
                  </button>
                </div>
              </div>
            )}
          </>
        )}
      </div>

      {/* Detail Modal */}
      {selectedNotification && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4"
          onClick={(e) => e.target === e.currentTarget && setSelectedNotification(null)}
        >
          <div className="card w-full max-w-lg">
            <div className="flex items-center justify-between border-b border-border px-6 py-4">
              <h3 className="text-lg font-semibold text-foreground">Notification Details</h3>
              <button
                onClick={() => setSelectedNotification(null)}
                className="text-muted-foreground hover:text-foreground"
              >
                <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth="2"
                    d="M6 18L18 6M6 6l12 12"
                  />
                </svg>
              </button>
            </div>
            <div className="space-y-3 p-6 text-sm">
              <div className="flex justify-between">
                <span className="text-muted-foreground">Type</span>
                <NotificationTypeBadge type={selectedNotification.notificationType} />
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Status</span>
                <NotificationStatusBadge status={selectedNotification.status} />
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Schedule Type</span>
                <span className="capitalize text-foreground">
                  {selectedNotification.scheduleType || '—'}
                </span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Recipient</span>
                <span className="capitalize text-foreground">
                  {selectedNotification.recipientType}
                </span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Scheduled For</span>
                <span className="text-foreground">
                  {selectedNotification.scheduledFor
                    ? formatDate(selectedNotification.scheduledFor)
                    : '—'}
                </span>
              </div>
              {selectedNotification.sentAt && (
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Sent At</span>
                  <span className="text-foreground">{formatDate(selectedNotification.sentAt)}</span>
                </div>
              )}
              <div>
                <span className="text-muted-foreground">Message</span>
                <p className="mt-1 rounded-lg bg-muted/50 p-3 text-foreground">
                  {selectedNotification.message}
                </p>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Debt Details Modal */}
      {selectedDebt && (
        <DebtDetailsModal
          debt={selectedDebt}
          onClose={() => setSelectedDebt(null)}
          onEdit={() => setSelectedDebt(null)}
          onDelete={() => setSelectedDebt(null)}
        />
      )}
    </div>
  )
}
