/**
 * Transforms a backend notification object (snake_case) to frontend format (camelCase)
 */
export const transformNotification = (notification) => {
  if (!notification) return null

  return {
    id: notification.id,
    debtListId: notification.debt_list_id,
    debtItemId: notification.debt_item_id || null,
    installmentNumber: notification.installment_number ?? null,
    installmentDueDate: notification.installment_due_date || null,
    notificationType: notification.notification_type,
    webhookType: notification.webhook_type || null,
    recipientType: notification.recipient_type,
    message: notification.message,
    status: notification.status,
    sentAt: notification.sent_at || null,
    scheduleType: notification.schedule_type || null,
    scheduledFor: notification.scheduled_for || null,
    cronJobId: notification.cron_job_id || null,
    reminderDaysBefore: notification.reminder_days_before ?? null,
    useCustomSchedule: notification.use_custom_schedule ?? false,
    customReminderDays: notification.custom_reminder_days || [],
    customNotificationTime: notification.custom_notification_time || null,
    customMessage: notification.custom_message || null,
    lastSentAt: notification.last_sent_at || null,
    nextRunAt: notification.next_run_at || null,
    isRecurring: notification.is_recurring ?? false,
    enabled: notification.enabled ?? true,
    recipientEmail: notification.recipient_email || null,
    recipientPhone: notification.recipient_phone || null,
    createdAt: notification.created_at || null,
    updatedAt: notification.updated_at || null,
  }
}

export const transformNotifications = (notifications) => {
  if (!Array.isArray(notifications)) return []
  return notifications.map(transformNotification).filter(Boolean)
}
