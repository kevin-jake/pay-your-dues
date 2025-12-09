import { useState, useEffect, useCallback } from 'react'
import { apiClient } from '@api/client'

export const TelegramSubscription = ({ telegramChatId, onStatusChange, disabled }) => {
  const [status, setStatus] = useState({
    configured: false,
    botUsername: '',
    botLink: '',
  })
  const [linkCode, setLinkCode] = useState(null)
  const [linkCodeExpiry, setLinkCodeExpiry] = useState(null)
  const [isLoading, setIsLoading] = useState(false)
  const [isUnlinking, setIsUnlinking] = useState(false)
  const [error, setError] = useState(null)
  const [copySuccess, setCopySuccess] = useState(false)

  // Check if Telegram is linked
  const isLinked = !!telegramChatId

  // Fetch Telegram bot status
  const fetchTelegramStatus = useCallback(async () => {
    try {
      const response = await apiClient.getTelegramStatus()
      setStatus({
        configured: response.configured,
        botUsername: response.bot_username || '',
        botLink: response.bot_link || '',
      })
    } catch (err) {
      console.error('Failed to fetch Telegram status:', err)
      setStatus({
        configured: false,
        botUsername: '',
        botLink: '',
      })
    }
  }, [])

  useEffect(() => {
    fetchTelegramStatus()
  }, [fetchTelegramStatus])

  // Countdown timer for link code expiry
  useEffect(() => {
    if (!linkCodeExpiry) return

    const interval = setInterval(() => {
      const now = new Date()
      const expiry = new Date(linkCodeExpiry)
      const remaining = Math.max(0, Math.floor((expiry - now) / 1000))

      if (remaining <= 0) {
        setLinkCode(null)
        setLinkCodeExpiry(null)
        clearInterval(interval)
      }
    }, 1000)

    return () => clearInterval(interval)
  }, [linkCodeExpiry])

  // Generate a new link code
  const handleGenerateLinkCode = async () => {
    setIsLoading(true)
    setError(null)

    try {
      const response = await apiClient.generateTelegramLinkCode()
      setLinkCode(response.code)
      setLinkCodeExpiry(new Date(Date.now() + response.expires_in * 1000))

      // Update status with bot info from response
      if (response.bot_username) {
        setStatus((prev) => ({
          ...prev,
          botUsername: response.bot_username,
          botLink: response.bot_link,
        }))
      }
    } catch (err) {
      setError(err.message || 'Failed to generate link code')
    } finally {
      setIsLoading(false)
    }
  }

  // Unlink Telegram account
  const handleUnlink = async () => {
    if (!confirm('Are you sure you want to unlink your Telegram account?')) {
      return
    }

    setIsUnlinking(true)
    setError(null)

    try {
      await apiClient.unlinkTelegram()
      setLinkCode(null)
      setLinkCodeExpiry(null)
      // Notify parent component to refresh settings
      if (onStatusChange) {
        onStatusChange()
      }
    } catch (err) {
      setError(err.message || 'Failed to unlink Telegram')
    } finally {
      setIsUnlinking(false)
    }
  }

  // Copy code to clipboard
  const handleCopyCode = async () => {
    if (!linkCode) return

    try {
      await navigator.clipboard.writeText(linkCode)
      setCopySuccess(true)
      setTimeout(() => setCopySuccess(false), 2000)
    } catch (err) {
      console.error('Failed to copy:', err)
    }
  }

  // Calculate remaining time
  const getRemainingTime = () => {
    if (!linkCodeExpiry) return null
    const now = new Date()
    const expiry = new Date(linkCodeExpiry)
    const remaining = Math.max(0, Math.floor((expiry - now) / 1000))
    const minutes = Math.floor(remaining / 60)
    const seconds = remaining % 60
    return `${minutes}:${seconds.toString().padStart(2, '0')}`
  }

  // Bot not configured at app level
  if (!status.configured) {
    return (
      <div className="rounded-lg border border-border bg-muted/30 p-4">
        <div className="flex items-start gap-3">
          <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-muted">
            <svg className="h-5 w-5 text-muted-foreground" viewBox="0 0 24 24" fill="currentColor">
              <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z" />
            </svg>
          </div>
          <div>
            <h4 className="font-medium text-foreground">Telegram Not Available</h4>
            <p className="mt-1 text-sm text-muted-foreground">
              Telegram notifications are not configured for this application. Please contact support
              if you'd like to use this feature.
            </p>
          </div>
        </div>
      </div>
    )
  }

  // Already linked
  if (isLinked) {
    return (
      <div className="rounded-lg border border-primary/20 bg-primary/5 p-4">
        <div className="flex items-start justify-between gap-4">
          <div className="flex items-start gap-3">
            <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-primary/10">
              <svg className="h-5 w-5 text-primary" viewBox="0 0 24 24" fill="currentColor">
                <path d="M9.78 18.65l.28-4.23 7.68-6.92c.34-.31-.07-.46-.52-.19L7.74 13.3 3.64 12c-.88-.25-.89-.86.2-1.3l15.97-6.16c.73-.33 1.43.18 1.15 1.3l-2.72 12.81c-.19.91-.74 1.13-1.5.71L12.6 16.3l-1.99 1.93c-.23.23-.42.42-.83.42z" />
              </svg>
            </div>
            <div>
              <h4 className="font-medium text-foreground">Telegram Connected</h4>
              <p className="mt-1 text-sm text-muted-foreground">
                Your Telegram account is linked. You will receive notifications via{' '}
                <a
                  href={status.botLink}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="font-medium text-primary hover:underline"
                >
                  @{status.botUsername}
                </a>
              </p>
              <p className="mt-1 text-xs text-muted-foreground">Chat ID: {telegramChatId}</p>
            </div>
          </div>
          <button
            onClick={handleUnlink}
            disabled={disabled || isUnlinking}
            className="shrink-0 rounded-md border border-destructive/30 bg-destructive/10 px-3 py-1.5 text-sm font-medium text-destructive transition-colors hover:bg-destructive/20 disabled:cursor-not-allowed disabled:opacity-50"
          >
            {isUnlinking ? 'Unlinking...' : 'Unlink'}
          </button>
        </div>
        {error && <p className="mt-2 text-sm text-destructive">{error}</p>}
      </div>
    )
  }

  // Not linked - show link flow
  return (
    <div className="rounded-lg border border-border p-4">
      <div className="flex items-start gap-3">
        <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-muted">
          <svg className="h-5 w-5 text-muted-foreground" viewBox="0 0 24 24" fill="currentColor">
            <path d="M9.78 18.65l.28-4.23 7.68-6.92c.34-.31-.07-.46-.52-.19L7.74 13.3 3.64 12c-.88-.25-.89-.86.2-1.3l15.97-6.16c.73-.33 1.43.18 1.15 1.3l-2.72 12.81c-.19.91-.74 1.13-1.5.71L12.6 16.3l-1.99 1.93c-.23.23-.42.42-.83.42z" />
          </svg>
        </div>
        <div className="flex-1">
          <h4 className="font-medium text-foreground">Connect Telegram</h4>
          <p className="mt-1 text-sm text-muted-foreground">
            Link your Telegram account to receive payment notifications.
          </p>

          {error && (
            <div className="mt-3 rounded-md bg-destructive/10 p-2 text-sm text-destructive">
              {error}
            </div>
          )}

          {!linkCode ? (
            <button
              onClick={handleGenerateLinkCode}
              disabled={disabled || isLoading}
              className="btn btn-primary mt-4"
            >
              {isLoading ? (
                <>
                  <svg className="mr-2 h-4 w-4 animate-spin" viewBox="0 0 24 24" fill="none">
                    <circle
                      className="opacity-25"
                      cx="12"
                      cy="12"
                      r="10"
                      stroke="currentColor"
                      strokeWidth="4"
                    />
                    <path
                      className="opacity-75"
                      fill="currentColor"
                      d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                    />
                  </svg>
                  Generating...
                </>
              ) : (
                'Generate Link Code'
              )}
            </button>
          ) : (
            <div className="mt-4 space-y-4">
              {/* Step 1: Open Bot */}
              <div className="flex items-start gap-3">
                <span className="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-primary text-xs font-bold text-primary-foreground">
                  1
                </span>
                <div>
                  <p className="text-sm font-medium text-foreground">Open the Telegram bot</p>
                  <a
                    href={status.botLink}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="mt-1 inline-flex items-center gap-1 text-sm text-primary hover:underline"
                  >
                    @{status.botUsername}
                    <svg className="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
                      <path
                        fillRule="evenodd"
                        d="M5.22 14.78a.75.75 0 001.06 0l7.22-7.22v5.69a.75.75 0 001.5 0v-7.5a.75.75 0 00-.75-.75h-7.5a.75.75 0 000 1.5h5.69l-7.22 7.22a.75.75 0 000 1.06z"
                        clipRule="evenodd"
                      />
                    </svg>
                  </a>
                </div>
              </div>

              {/* Step 2: Send Code */}
              <div className="flex items-start gap-3">
                <span className="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-primary text-xs font-bold text-primary-foreground">
                  2
                </span>
                <div className="flex-1">
                  <p className="text-sm font-medium text-foreground">Send this code to the bot</p>
                  <div className="mt-2 flex items-center gap-2">
                    <code className="rounded-md bg-muted px-3 py-2 font-mono text-lg font-bold tracking-widest text-foreground">
                      {linkCode}
                    </code>
                    <button
                      onClick={handleCopyCode}
                      className="rounded-md border border-border p-2 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
                      title="Copy code"
                    >
                      {copySuccess ? (
                        <svg className="h-5 w-5 text-green-500" viewBox="0 0 20 20" fill="currentColor">
                          <path
                            fillRule="evenodd"
                            d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"
                            clipRule="evenodd"
                          />
                        </svg>
                      ) : (
                        <svg className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                          <path d="M8 3a1 1 0 011-1h2a1 1 0 110 2H9a1 1 0 01-1-1z" />
                          <path d="M6 3a2 2 0 00-2 2v11a2 2 0 002 2h8a2 2 0 002-2V5a2 2 0 00-2-2 3 3 0 01-3 3H9a3 3 0 01-3-3z" />
                        </svg>
                      )}
                    </button>
                  </div>
                  <p className="mt-1 text-xs text-muted-foreground">
                    Code expires in{' '}
                    <span className="font-medium text-foreground">{getRemainingTime()}</span>
                  </p>
                </div>
              </div>

              {/* Step 3: Wait */}
              <div className="flex items-start gap-3">
                <span className="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-muted text-xs font-bold text-muted-foreground">
                  3
                </span>
                <div>
                  <p className="text-sm font-medium text-muted-foreground">
                    Refresh this page after linking
                  </p>
                </div>
              </div>

              <div className="flex items-center gap-2 pt-2">
                <button
                  onClick={handleGenerateLinkCode}
                  disabled={disabled || isLoading}
                  className="text-sm text-muted-foreground hover:text-foreground"
                >
                  Generate new code
                </button>
                <span className="text-muted-foreground">·</span>
                <button
                  onClick={() => {
                    setLinkCode(null)
                    setLinkCodeExpiry(null)
                  }}
                  className="text-sm text-muted-foreground hover:text-foreground"
                >
                  Cancel
                </button>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

