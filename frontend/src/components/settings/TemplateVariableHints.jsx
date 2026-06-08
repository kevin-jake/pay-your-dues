const TEMPLATE_VARIABLES = [
  { name: '{{recipient_name}}', description: 'Name of the notification recipient' },
  { name: '{{contact_name}}', description: 'Contact name on the debt' },
  { name: '{{amount}}', description: 'Payment or installment amount' },
  { name: '{{currency}}', description: 'Currency code' },
  { name: '{{due_date}}', description: 'Due date formatted' },
  { name: '{{installment_number}}', description: 'Current installment number' },
  { name: '{{installment_total}}', description: 'Total number of installments' },
  { name: '{{days_until_due}}', description: 'Days until due date' },
  { name: '{{remaining_debt}}', description: 'Remaining debt balance' },
  { name: '{{payment_amount}}', description: 'Payment amount (event notifications)' },
]

export const TemplateVariableHints = () => {
  return (
    <div className="mt-2 rounded-lg bg-muted/50 p-3">
      <p className="mb-2 text-xs font-medium text-foreground">Available template variables:</p>
      <div className="flex flex-wrap gap-1">
        {TEMPLATE_VARIABLES.map((variable) => (
          <span
            key={variable.name}
            title={variable.description}
            className="cursor-help rounded bg-background px-1.5 py-0.5 font-mono text-xs text-muted-foreground"
          >
            {variable.name}
          </span>
        ))}
      </div>
      <p className="mt-2 text-xs text-muted-foreground">
        Use these placeholders in your custom message. They are replaced with actual values when
        notifications are sent.
      </p>
    </div>
  )
}
