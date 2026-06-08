import { useState } from 'react'
import {
  NotificationStatusBadge,
  NotificationTypeBadge,
} from '@components/notifications/NotificationStatusBadge'
import { formatDate } from '@utils/formatters'

export const NotificationScheduleRow = ({ group, onDelete, onToggleChannel, isActionLoading }) => {
  const [expanded, setExpanded] = useState(false)

  const overallStatus = deriveGroupStatus(group.channels)

  return (
    <>
      {/* Collapsed row */}
      <tr className="border-b border-border/50 hover:bg-muted/20">
        <td className="py-3 pl-2 pr-4">
          <button
            onClick={() => setExpanded((v) => !v)}
            className="flex items-center gap-1 text-muted-foreground hover:text-foreground"
            aria-label={expanded ? 'Collapse' : 'Expand'}
          >
            <svg
              className={`h-4 w-4 transition-transform ${expanded ? 'rotate-90' : ''}`}
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9 5l7 7-7 7" />
            </svg>
          </button>
        </td>
        <td className="py-3 pr-4 text-sm text-foreground">
          {group.installmentNumber != null ? `#${group.installmentNumber}` : '—'}
        </td>
        <td className="py-3 pr-4 text-sm text-foreground">
          {group.scheduledFor ? formatDate(group.scheduledFor) : '—'}
        </td>
        <td className="py-3 pr-4">
          <NotificationStatusBadge status={overallStatus} />
        </td>
        <td className="py-3 pr-4 text-xs text-muted-foreground">
          {group.channels.length} channel{group.channels.length !== 1 ? 's' : ''}
        </td>
        <td className="py-3">
          <button
            onClick={() =>
              onDelete({
                installmentNumber: group.installmentNumber,
                scheduledFor: group.scheduledFor,
              })
            }
            disabled={isActionLoading}
            className="text-xs text-destructive hover:underline disabled:cursor-not-allowed disabled:opacity-50"
          >
            Delete
          </button>
        </td>
      </tr>

      {/* Expanded channels */}
      {expanded &&
        group.channels.map((n) => (
          <tr key={n.id} className="border-b border-border/30 bg-muted/10">
            <td className="py-2 pl-8 pr-4" />
            <td className="py-2 pr-4">
              <NotificationTypeBadge type={n.notificationType} />
            </td>
            <td className="py-2 pr-4 text-xs text-muted-foreground">
              {n.reminderDaysBefore != null ? `${n.reminderDaysBefore}d before` : '—'}
            </td>
            <td className="py-2 pr-4">
              <NotificationStatusBadge status={n.status} />
            </td>
            <td className="py-2 pr-4">
              <button
                onClick={() => onToggleChannel(n)}
                disabled={isActionLoading}
                className={`relative inline-flex h-5 w-9 items-center rounded-full transition-colors disabled:cursor-not-allowed disabled:opacity-50 ${
                  n.enabled ? 'bg-primary' : 'bg-muted'
                }`}
                aria-label={n.enabled ? 'Disable channel' : 'Enable channel'}
              >
                <span
                  className={`inline-block h-3 w-3 transform rounded-full bg-white transition-transform ${
                    n.enabled ? 'translate-x-5' : 'translate-x-1'
                  }`}
                />
              </button>
            </td>
            <td />
          </tr>
        ))}
    </>
  )
}

function deriveGroupStatus(channels) {
  if (channels.every((c) => c.status === 'sent')) return 'sent'
  if (channels.some((c) => c.status === 'failed')) return 'failed'
  if (channels.some((c) => c.status === 'queued')) return 'queued'
  if (channels.every((c) => ['skipped', 'cancelled'].includes(c.status))) return 'cancelled'
  return 'pending'
}
