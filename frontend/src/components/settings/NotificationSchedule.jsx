import { useState } from 'react'

export const NotificationSchedule = ({ advancedInputs, onInputChange, isSaving }) => {
  const [showNotificationSchedule, setShowNotificationSchedule] = useState(false)

  return (
    <div className="mt-6 border-t border-border pt-6">
      <button
        onClick={() => setShowNotificationSchedule(!showNotificationSchedule)}
        className="flex w-full items-center justify-between text-left"
      >
        <h3 className="text-lg font-semibold text-foreground">Notification Schedule</h3>
        <svg
          className={`h-5 w-5 text-muted-foreground transition-transform ${
            showNotificationSchedule ? 'rotate-180' : ''
          }`}
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M19 9l-7 7-7-7" />
        </svg>
      </button>

      {showNotificationSchedule && (
        <div className="mt-4 space-y-6">
          {/* Reminder Schedule */}
          <div>
            <label className="mb-2 block text-sm font-medium text-foreground">Reminder Days</label>
            <div className="flex flex-wrap gap-2">
              {[7, 3, 1].map((day) => (
                <button
                  key={day}
                  onClick={() => {
                    const currentDays = advancedInputs.notificationReminderDays || []
                    const newDays = currentDays.includes(day)
                      ? currentDays.filter((d) => d !== day)
                      : [...currentDays, day].sort((a, b) => b - a)
                    onInputChange('notificationReminderDays', newDays)
                  }}
                  disabled={isSaving}
                  className={`rounded-lg border px-3 py-1 text-sm transition-colors disabled:cursor-not-allowed disabled:opacity-50 ${
                    advancedInputs.notificationReminderDays?.includes(day)
                      ? 'border-primary bg-primary/10 text-primary'
                      : 'border-border hover:border-primary/50'
                  }`}
                >
                  {day} {day === 1 ? 'day' : 'days'} before
                </button>
              ))}
            </div>
            <p className="mt-1 text-xs text-muted-foreground">
              Select when to receive reminders before due dates
            </p>
          </div>

          {/* Notification Time */}
          <div>
            <label className="mb-2 block text-sm font-medium text-foreground">
              Notification Time
            </label>
            <input
              type="time"
              value={advancedInputs.notificationTime}
              onChange={(e) => onInputChange('notificationTime', e.target.value)}
              disabled={isSaving}
              className="input max-w-xs disabled:cursor-not-allowed disabled:opacity-50"
            />
            <p className="mt-1 text-xs text-muted-foreground">
              Time of day to send notifications (24-hour format)
            </p>
          </div>

          {/* Overdue Reminder Frequency */}
          <div>
            <label className="mb-2 block text-sm font-medium text-foreground">
              Overdue Reminder Frequency
            </label>
            <select
              value={advancedInputs.overdueReminderFrequency}
              onChange={(e) => onInputChange('overdueReminderFrequency', e.target.value)}
              disabled={isSaving}
              className="input max-w-xs disabled:cursor-not-allowed disabled:opacity-50"
            >
              <option value="daily">Daily</option>
              <option value="weekly">Weekly</option>
              <option value="monthly">Monthly</option>
            </select>
            <p className="mt-1 text-xs text-muted-foreground">
              How often to send reminders for overdue payments
            </p>
          </div>
        </div>
      )}
    </div>
  )
}
