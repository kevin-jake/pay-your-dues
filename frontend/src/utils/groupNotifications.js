/**
 * Groups flat notification records into slot groups:
 * one group per (installmentNumber, scheduledFor day).
 * Only 'reminder' and 'overdue' schedule types are included.
 *
 * @param {Array} notifications - transformed notification objects (camelCase)
 * @returns {Array} sorted array of slot groups
 */
export const groupNotificationsBySlot = (notifications) => {
  const groups = {}

  for (const n of notifications) {
    if (!['reminder', 'overdue'].includes(n.scheduleType)) continue

    const day = n.scheduledFor
      ? new Date(n.scheduledFor).toISOString().split('T')[0]
      : 'unscheduled'
    const installmentKey = n.installmentNumber != null ? n.installmentNumber : 'onetime'
    const key = `${installmentKey}|${day}`

    if (!groups[key]) {
      groups[key] = {
        key,
        installmentNumber: n.installmentNumber,
        installmentDueDate: n.installmentDueDate,
        scheduledFor: n.scheduledFor,
        reminderDaysBefore: n.reminderDaysBefore,
        channels: [],
      }
    }
    groups[key].channels.push(n)
  }

  return Object.values(groups).sort((a, b) => {
    const dateA = a.scheduledFor ? new Date(a.scheduledFor) : new Date(0)
    const dateB = b.scheduledFor ? new Date(b.scheduledFor) : new Date(0)
    return dateA - dateB
  })
}
