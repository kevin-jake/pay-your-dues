import { useState } from 'react'
import { NotificationToggle } from './NotificationToggle'
import { TelegramSubscription } from './TelegramSubscription'

export const WebhookConfiguration = ({
  webhookInputs,
  webhookEnabled,
  onWebhookToggle,
  onInputChange,
  onRefreshSettings,
  isSaving,
  validationErrors,
}) => {
  const [showWebhookConfig, setShowWebhookConfig] = useState(false)

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
        <div className="mt-4">
          {/* Webhook Notifications Master Toggle */}
          <NotificationToggle
            label="Enable Webhook Notifications"
            description="Send notifications to configured webhooks (Slack, Telegram, Discord)"
            enabled={webhookEnabled}
            onChange={onWebhookToggle}
            disabled={isSaving}
          />
        </div>
      )}

      {showWebhookConfig && webhookEnabled && (
        <div className="mt-4 space-y-6 border-t border-border pt-4">
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
              className={`input disabled:cursor-not-allowed disabled:opacity-50 ${
                validationErrors?.slack_webhook_url ? 'border-destructive' : ''
              }`}
            />
            {validationErrors?.slack_webhook_url && (
              <p className="mt-1 text-xs text-destructive">{validationErrors.slack_webhook_url}</p>
            )}
            <p className="mt-1 text-xs text-muted-foreground">
              Enter your Slack webhook URL to receive notifications
            </p>
          </div>

          {/* Telegram Configuration - New Subscription Flow */}
          <div>
            <label className="mb-2 block text-sm font-medium text-foreground">
              Telegram Notifications
            </label>
            <TelegramSubscription
              telegramChatId={webhookInputs.telegramChatId}
              onStatusChange={onRefreshSettings}
              disabled={isSaving}
            />
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
              className={`input disabled:cursor-not-allowed disabled:opacity-50 ${
                validationErrors?.discord_webhook_url ? 'border-destructive' : ''
              }`}
            />
            {validationErrors?.discord_webhook_url && (
              <p className="mt-1 text-xs text-destructive">
                {validationErrors.discord_webhook_url}
              </p>
            )}
            <p className="mt-1 text-xs text-muted-foreground">
              Enter your Discord webhook URL to receive notifications
            </p>
          </div>
        </div>
      )}
    </div>
  )
}
