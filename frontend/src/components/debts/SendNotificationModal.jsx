import { useState } from 'react'

const MAX_MANUAL = 3

export const SendNotificationModal = ({ debt, onClose, onSend, isLoading, sendLimits }) => {
  const [message, setMessage] = useState('')
  const [notificationType, setNotificationType] = useState('email')

  const getRemaining = (type) => {
    if (!sendLimits) return MAX_MANUAL
    return type === 'email' ? (sendLimits.email?.remaining ?? MAX_MANUAL) : (sendLimits.sms?.remaining ?? MAX_MANUAL)
  }

  const remaining = getRemaining(notificationType)
  const isLimitReached = remaining <= 0

  const handleSubmit = async (e) => {
    e.preventDefault()
    if (!message.trim() || isLimitReached) return
    await onSend({ message: message.trim(), notificationType })
  }

  const handleOverlayClick = (e) => {
    if (e.target === e.currentTarget) {
      onClose()
    }
  }

  return (
    <div
      className="fixed inset-0 z-[60] flex items-center justify-center bg-black/60 p-4"
      onClick={handleOverlayClick}
    >
      <div className="card w-full max-w-md">
        <div className="border-b border-border px-6 py-4">
          <h3 className="text-lg font-semibold text-foreground">Send Notification</h3>
          <p className="mt-1 text-sm text-muted-foreground">
            Send an immediate notification for {debt.contact?.name || 'this debt'}
          </p>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4 p-6">
          <div>
            <label className="mb-2 block text-sm font-medium text-foreground">Channel</label>
            <select
              value={notificationType}
              onChange={(e) => setNotificationType(e.target.value)}
              disabled={isLoading}
              className="input disabled:cursor-not-allowed disabled:opacity-50"
            >
              {['email', 'sms'].map((type) => {
                const rem = getRemaining(type)
                return (
                  <option key={type} value={type} disabled={rem <= 0}>
                    {type.toUpperCase()} — {rem}/{MAX_MANUAL} sends remaining
                  </option>
                )
              })}
            </select>
          </div>

          {sendLimits && (
            <div className="flex gap-4 rounded-lg bg-muted/40 px-3 py-2 text-xs text-muted-foreground">
              <span>Email: {sendLimits.email?.remaining ?? MAX_MANUAL}/{MAX_MANUAL} left</span>
              <span>SMS: {sendLimits.sms?.remaining ?? MAX_MANUAL}/{MAX_MANUAL} left</span>
            </div>
          )}

          {isLimitReached && (
            <p className="rounded-lg bg-destructive/10 px-3 py-2 text-xs text-destructive">
              You have reached the maximum of {MAX_MANUAL} manual sends for {notificationType.toUpperCase()} on this debt.
            </p>
          )}

          <div>
            <label className="mb-2 block text-sm font-medium text-foreground">Message</label>
            <textarea
              value={message}
              onChange={(e) => setMessage(e.target.value)}
              placeholder="Enter your message..."
              rows={4}
              required
              disabled={isLoading || isLimitReached}
              className="input disabled:cursor-not-allowed disabled:opacity-50"
            />
          </div>

          <div className="flex space-x-3">
            <button
              type="button"
              onClick={onClose}
              disabled={isLoading}
              className="btn-secondary flex-1 disabled:cursor-not-allowed disabled:opacity-50"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={isLoading || !message.trim() || isLimitReached}
              className="btn-primary flex-1 disabled:cursor-not-allowed disabled:opacity-50"
            >
              {isLoading ? 'Sending…' : 'Send Now'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
