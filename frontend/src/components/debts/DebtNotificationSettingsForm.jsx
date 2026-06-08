import { useState } from 'react'

const REMINDER_DAY_OPTIONS = [1, 3, 7, 14]

export const DebtNotificationSettingsForm = ({
  initialSettings,
  onSave,
  onCancel,
  isSaving,
}) => {
  const [settings, setSettings] = useState({
    reminder_days: initialSettings?.reminder_days ?? [7, 3, 1],
    notification_time: initialSettings?.notification_time ?? '09:00',
    notify_email: initialSettings?.notify_email ?? true,
    notify_sms: initialSettings?.notify_sms ?? false,
    notify_slack: initialSettings?.notify_slack ?? false,
    notify_telegram: initialSettings?.notify_telegram ?? false,
    notify_discord: initialSettings?.notify_discord ?? false,
  })

  const toggleReminderDay = (day) => {
    setSettings((prev) => {
      const days = prev.reminder_days.includes(day)
        ? prev.reminder_days.filter((d) => d !== day)
        : [...prev.reminder_days, day].sort((a, b) => b - a)
      return { ...prev, reminder_days: days }
    })
  }

  const handleSubmit = (e) => {
    e.preventDefault()
    onSave(settings)
  }

  const hasAnyChannel =
    settings.notify_email ||
    settings.notify_sms ||
    settings.notify_slack ||
    settings.notify_telegram ||
    settings.notify_discord

  return (
    <form onSubmit={handleSubmit} className="space-y-5 rounded-lg border border-border bg-muted/30 p-4">
      <div>
        <label className="mb-2 block text-sm font-medium text-foreground">Reminder Days Before Due</label>
        <div className="flex flex-wrap gap-2">
          {REMINDER_DAY_OPTIONS.map((day) => (
            <button
              key={day}
              type="button"
              onClick={() => toggleReminderDay(day)}
              disabled={isSaving}
              className={`rounded-lg border px-3 py-1 text-sm transition-colors disabled:cursor-not-allowed disabled:opacity-50 ${
                settings.reminder_days.includes(day)
                  ? 'border-primary bg-primary/10 text-primary'
                  : 'border-border hover:border-primary/50 text-muted-foreground'
              }`}
            >
              {day} {day === 1 ? 'day' : 'days'} before
            </button>
          ))}
        </div>
        {settings.reminder_days.length === 0 && (
          <p className="mt-1 text-xs text-destructive">Select at least one reminder day</p>
        )}
      </div>

      <div>
        <label className="mb-2 block text-sm font-medium text-foreground">Send Time</label>
        <input
          type="time"
          value={settings.notification_time}
          onChange={(e) => setSettings((p) => ({ ...p, notification_time: e.target.value }))}
          disabled={isSaving}
          className="input max-w-xs disabled:cursor-not-allowed disabled:opacity-50"
        />
        <p className="mt-1 text-xs text-muted-foreground">Time of day to deliver reminders (24 h)</p>
      </div>

      <div>
        <label className="mb-2 block text-sm font-medium text-foreground">Channels</label>
        <div className="space-y-2">
          {[
            { key: 'notify_email', label: 'Email' },
            { key: 'notify_sms', label: 'SMS' },
            { key: 'notify_slack', label: 'Slack' },
            { key: 'notify_telegram', label: 'Telegram' },
            { key: 'notify_discord', label: 'Discord' },
          ].map(({ key, label }) => (
            <label key={key} className="flex cursor-pointer items-center gap-3">
              <input
                type="checkbox"
                checked={settings[key]}
                onChange={(e) => setSettings((p) => ({ ...p, [key]: e.target.checked }))}
                disabled={isSaving}
                className="h-4 w-4 accent-primary disabled:cursor-not-allowed"
              />
              <span className="text-sm text-foreground">{label}</span>
            </label>
          ))}
        </div>
        {!hasAnyChannel && (
          <p className="mt-1 text-xs text-destructive">Enable at least one channel</p>
        )}
      </div>

      <div className="flex gap-2 pt-1">
        <button
          type="button"
          onClick={onCancel}
          disabled={isSaving}
          className="btn-secondary flex-1 disabled:cursor-not-allowed disabled:opacity-50"
        >
          Cancel
        </button>
        <button
          type="submit"
          disabled={isSaving || settings.reminder_days.length === 0 || !hasAnyChannel}
          className="btn-primary flex-1 disabled:cursor-not-allowed disabled:opacity-50"
        >
          {isSaving ? 'Saving…' : 'Enable Reminders'}
        </button>
      </div>
    </form>
  )
}
