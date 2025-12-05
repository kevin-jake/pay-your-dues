import { useState } from 'react'

export const WebhookConfiguration = ({ webhookInputs, onInputChange, isSaving }) => {
  const [showWebhookConfig, setShowWebhookConfig] = useState(false)
  const [showToken, setShowToken] = useState(false)

  return (
    <div className="mt-6 border-t border-border pt-6">
      <button
        onClick={() => setShowWebhookConfig(!showWebhookConfig)}
        className="flex w-full items-center justify-between text-left"
      >
        <h3 className="text-lg font-semibold text-foreground">Webhook Configuration</h3>
        <svg
          className={`h-5 w-5 text-muted-foreground transition-transform ${
            showWebhookConfig ? 'rotate-180' : ''
          }`}
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M19 9l-7 7-7-7" />
        </svg>
      </button>

      {showWebhookConfig && (
        <div className="mt-4 space-y-4">
          {/* Slack Webhook */}
          <div>
            <label className="mb-2 block text-sm font-medium text-foreground">
              Slack Webhook URL
            </label>
            <input
              type="url"
              value={webhookInputs.slackWebhookUrl}
              onChange={(e) => onInputChange('slackWebhookUrl', e.target.value)}
              placeholder="https://hooks.slack.com/services/..."
              disabled={isSaving}
              className="input disabled:cursor-not-allowed disabled:opacity-50"
            />
            <p className="mt-1 text-xs text-muted-foreground">
              Enter your Slack webhook URL to receive notifications
            </p>
          </div>

          {/* Telegram Configuration */}
          <div>
            <label className="mb-2 block text-sm font-medium text-foreground">
              Telegram Bot Token
            </label>
            <div className="relative">
              <input
                type={showToken ? 'text' : 'password'}
                value={webhookInputs.telegramBotToken}
                onChange={(e) => onInputChange('telegramBotToken', e.target.value)}
                placeholder="Enter bot token"
                disabled={isSaving}
                className="input pr-10 disabled:cursor-not-allowed disabled:opacity-50"
              />
              <button
                type="button"
                onClick={() => setShowToken(!showToken)}
                disabled={isSaving}
                className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground transition-colors hover:text-foreground disabled:cursor-not-allowed disabled:opacity-50"
                aria-label={showToken ? 'Hide token' : 'Show token'}
              >
                {showToken ? (
                  <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth="2"
                      d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21"
                    />
                  </svg>
                ) : (
                  <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth="2"
                      d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
                    />
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth="2"
                      d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"
                    />
                  </svg>
                )}
              </button>
            </div>
            <p className="mt-1 text-xs text-muted-foreground">
              Your Telegram bot token (kept secure)
            </p>
          </div>

          <div>
            <label className="mb-2 block text-sm font-medium text-foreground">
              Telegram Chat ID
            </label>
            <input
              type="text"
              value={webhookInputs.telegramChatId}
              onChange={(e) => onInputChange('telegramChatId', e.target.value)}
              placeholder="Enter chat ID"
              disabled={isSaving}
              className="input disabled:cursor-not-allowed disabled:opacity-50"
            />
            <p className="mt-1 text-xs text-muted-foreground">
              Your Telegram chat ID for receiving notifications
            </p>
          </div>

          {/* Discord Webhook */}
          <div>
            <label className="mb-2 block text-sm font-medium text-foreground">
              Discord Webhook URL
            </label>
            <input
              type="url"
              value={webhookInputs.discordWebhookUrl}
              onChange={(e) => onInputChange('discordWebhookUrl', e.target.value)}
              placeholder="https://discord.com/api/webhooks/..."
              disabled={isSaving}
              className="input disabled:cursor-not-allowed disabled:opacity-50"
            />
            <p className="mt-1 text-xs text-muted-foreground">
              Enter your Discord webhook URL to receive notifications
            </p>
          </div>
        </div>
      )}
    </div>
  )
}
