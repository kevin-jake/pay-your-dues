package notification

import (
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

// TemplateEngine handles template rendering with variable substitution
type TemplateEngine struct{}

// NewTemplateEngine creates a new template engine instance
func NewTemplateEngine() *TemplateEngine {
	return &TemplateEngine{}
}

// TemplateData contains all data for template variable substitution
type TemplateData struct {
	// Recipient type ('user' or 'contact') - determines who receives the notification
	RecipientType string

	// User information
	UserFirstName string
	UserLastName  string
	UserEmail     string

	// Contact information
	ContactName  string
	ContactEmail string
	ContactPhone string

	// Debt information
	Amount           decimal.Decimal
	Currency         string
	DueDate          time.Time
	TotalDebt        decimal.Decimal
	RemainingDebt    decimal.Decimal
	PaidAmount       decimal.Decimal
	DebtType         string
	InstallmentPlan  string

	// Installment-specific information
	InstallmentNumber       int
	InstallmentTotal        int
	InstallmentDueDate      time.Time
	InstallmentAmount       decimal.Decimal
	RemainingInstallments   int
	PaymentsMadeCount       int
	DaysUntilDue            int

	// Payment information (for event notifications)
	PaymentAmount     decimal.Decimal
	PaymentDate       time.Time
	PaymentMethod     string
	PaymentStatus     string
	VerificationStatus string
	RejectionReason   string

	// Custom fields
	CustomMessage string
}

// Render renders a template with the provided data
func (te *TemplateEngine) Render(template string, data TemplateData) string {
	result := template

	// Determine recipient-aware names based on RecipientType
	// When recipient is 'contact', the notification is addressed to the contact
	// When recipient is 'user', the notification is addressed to the user
	recipientName := data.UserFirstName
	recipientFullName := fmt.Sprintf("%s %s", data.UserFirstName, data.UserLastName)
	otherPartyName := data.ContactName
	if data.RecipientType == "contact" {
		recipientName = data.ContactName
		recipientFullName = data.ContactName
		otherPartyName = fmt.Sprintf("%s %s", data.UserFirstName, data.UserLastName)
	}

	// Replace recipient-aware variables (use these in templates for proper addressing)
	result = strings.ReplaceAll(result, "{{recipient_name}}", recipientName)
	result = strings.ReplaceAll(result, "{{recipient_full_name}}", recipientFullName)
	result = strings.ReplaceAll(result, "{{other_party_name}}", otherPartyName)

	// Replace user variables
	result = strings.ReplaceAll(result, "{{user_first_name}}", data.UserFirstName)
	result = strings.ReplaceAll(result, "{{user_last_name}}", data.UserLastName)
	result = strings.ReplaceAll(result, "{{user_full_name}}", fmt.Sprintf("%s %s", data.UserFirstName, data.UserLastName))
	result = strings.ReplaceAll(result, "{{user_email}}", data.UserEmail)

	// Replace contact variables
	result = strings.ReplaceAll(result, "{{contact_name}}", data.ContactName)
	result = strings.ReplaceAll(result, "{{debtor_name}}", data.ContactName) // Alias for contact_name
	result = strings.ReplaceAll(result, "{{contact_email}}", data.ContactEmail)
	result = strings.ReplaceAll(result, "{{contact_phone}}", data.ContactPhone)

	// Replace debt variables
	result = strings.ReplaceAll(result, "{{amount}}", data.Amount.StringFixed(2))
	result = strings.ReplaceAll(result, "{{currency}}", data.Currency)
	result = strings.ReplaceAll(result, "{{due_date}}", data.DueDate.Format("January 02, 2006"))
	result = strings.ReplaceAll(result, "{{total_debt}}", data.TotalDebt.StringFixed(2))
	result = strings.ReplaceAll(result, "{{remaining_debt}}", data.RemainingDebt.StringFixed(2))
	result = strings.ReplaceAll(result, "{{paid_amount}}", data.PaidAmount.StringFixed(2))
	result = strings.ReplaceAll(result, "{{debt_type}}", data.DebtType)
	result = strings.ReplaceAll(result, "{{installment_plan}}", data.InstallmentPlan)

	// Replace installment-specific variables
	result = strings.ReplaceAll(result, "{{installment_number}}", fmt.Sprintf("%d", data.InstallmentNumber))
	result = strings.ReplaceAll(result, "{{installment_total}}", fmt.Sprintf("%d", data.InstallmentTotal))
	result = strings.ReplaceAll(result, "{{installment_due_date}}", data.InstallmentDueDate.Format("January 02, 2006"))
	result = strings.ReplaceAll(result, "{{installment_amount}}", data.InstallmentAmount.StringFixed(2))
	result = strings.ReplaceAll(result, "{{remaining_installments}}", fmt.Sprintf("%d", data.RemainingInstallments))
	result = strings.ReplaceAll(result, "{{payments_made_count}}", fmt.Sprintf("%d", data.PaymentsMadeCount))
	result = strings.ReplaceAll(result, "{{days_until_due}}", fmt.Sprintf("%d", data.DaysUntilDue))

	// Replace payment variables
	result = strings.ReplaceAll(result, "{{payment_amount}}", data.PaymentAmount.StringFixed(2))
	result = strings.ReplaceAll(result, "{{payment_date}}", data.PaymentDate.Format("January 02, 2006"))
	result = strings.ReplaceAll(result, "{{payment_method}}", data.PaymentMethod)
	result = strings.ReplaceAll(result, "{{payment_status}}", data.PaymentStatus)
	result = strings.ReplaceAll(result, "{{verification_status}}", data.VerificationStatus)
	result = strings.ReplaceAll(result, "{{rejection_reason}}", data.RejectionReason)

	// Replace custom message
	result = strings.ReplaceAll(result, "{{custom_message}}", data.CustomMessage)

	return result
}

// RenderHTML renders an HTML template with the provided data
func (te *TemplateEngine) RenderHTML(template string, data TemplateData) string {
	// HTML rendering uses the same variable substitution as plain text
	return te.Render(template, data)
}

// GetDefaultEmailTemplate returns a default email template for payment reminders
func GetDefaultEmailTemplate() string {
	return `<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #4caf50; color: white; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
        .content { padding: 20px; background-color: #f9f9f9; }
        .details { background-color: white; padding: 15px; margin: 10px 0; border-left: 4px solid #4caf50; border-radius: 3px; }
        .footer { padding: 20px; text-align: center; font-size: 12px; color: #666; }
        .amount { font-size: 24px; font-weight: bold; color: #4caf50; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h2>Payment Reminder</h2>
        </div>
        <div class="content">
            <p>Hi {{recipient_name}},</p>
            <p>This is a friendly reminder that a payment is due to <strong>{{other_party_name}}</strong>.</p>
            
            <div class="details">
                <h3>Payment Details:</h3>
                <ul>
                    <li><strong>Amount:</strong> <span class="amount">{{currency}} {{amount}}</span></li>
                    <li><strong>Due Date:</strong> {{due_date}}</li>
                    <li><strong>Days Until Due:</strong> {{days_until_due}} days</li>
                </ul>
            </div>

            <p>Please make sure to complete this payment on time.</p>
            
            <p>Best regards,<br>Pay Your Dues Team</p>
        </div>
        <div class="footer">
            <p>This is an automated reminder. Please do not reply to this email.</p>
        </div>
    </div>
</body>
</html>`
}

// GetDefaultSMSTemplate returns a default SMS template for payment reminders
func GetDefaultSMSTemplate() string {
	return `Reminder: Payment of {{currency}} {{amount}} due to {{other_party_name}} on {{due_date}} ({{days_until_due}} days). - Pay Your Dues`
}

// GetInstallmentEmailTemplate returns a default email template for installment reminders
func GetInstallmentEmailTemplate() string {
	return `<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #4caf50; color: white; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
        .content { padding: 20px; background-color: #f9f9f9; }
        .details { background-color: white; padding: 15px; margin: 10px 0; border-left: 4px solid #4caf50; border-radius: 3px; }
        .footer { padding: 20px; text-align: center; font-size: 12px; color: #666; }
        .amount { font-size: 24px; font-weight: bold; color: #4caf50; }
        .installment-badge { display: inline-block; background-color: #2196F3; color: white; padding: 5px 10px; border-radius: 3px; font-weight: bold; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h2>Installment Payment Reminder</h2>
            <span class="installment-badge">Installment #{{installment_number}} of {{installment_total}}</span>
        </div>
        <div class="content">
            <p>Hi {{recipient_name}},</p>
            <p>This is a reminder that installment payment <strong>#{{installment_number}} of {{installment_total}}</strong> is due to <strong>{{other_party_name}}</strong> in <strong>{{days_until_due}} days</strong>.</p>
            
            <div class="details">
                <h3>Payment Details:</h3>
                <ul>
                    <li><strong>Installment:</strong> #{{installment_number}} of {{installment_total}}</li>
                    <li><strong>Amount:</strong> <span class="amount">{{currency}} {{installment_amount}}</span></li>
                    <li><strong>Due Date:</strong> {{installment_due_date}}</li>
                    <li><strong>Remaining Installments:</strong> {{remaining_installments}}</li>
                </ul>
            </div>

            <div class="details">
                <h3>Overall Progress:</h3>
                <ul>
                    <li><strong>Total Debt:</strong> {{currency}} {{total_debt}}</li>
                    <li><strong>Remaining Debt:</strong> {{currency}} {{remaining_debt}}</li>
                    <li><strong>Payments Made:</strong> {{payments_made_count}} of {{installment_total}}</li>
                </ul>
            </div>

            <p>Please make sure to complete this payment on time.</p>
            
            <p>Best regards,<br>Pay Your Dues Team</p>
        </div>
        <div class="footer">
            <p>This is an automated reminder. Please do not reply to this email.</p>
        </div>
    </div>
</body>
</html>`
}

// GetInstallmentSMSTemplate returns a default SMS template for installment reminders
func GetInstallmentSMSTemplate() string {
	return `Reminder: Installment #{{installment_number}}/{{installment_total}} of {{currency}} {{installment_amount}} due to {{other_party_name}} on {{installment_due_date}} ({{days_until_due}} days).`
}

// GetPaymentConfirmationEmailTemplate returns a template for payment confirmation notifications
func GetPaymentConfirmationEmailTemplate() string {
	return `<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #4caf50; color: white; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
        .content { padding: 20px; background-color: #f9f9f9; }
        .details { background-color: white; padding: 15px; margin: 10px 0; border-left: 4px solid #4caf50; border-radius: 3px; }
        .footer { padding: 20px; text-align: center; font-size: 12px; color: #666; }
        .status-pending { color: #ff9800; font-weight: bold; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h2>Payment Received</h2>
        </div>
        <div class="content">
            <p>Hi {{recipient_name}},</p>
            <p>A payment has been recorded for your debt with <strong>{{other_party_name}}</strong>.</p>
            
            <div class="details">
                <h3>Payment Information:</h3>
                <ul>
                    <li><strong>Amount:</strong> {{currency}} {{payment_amount}}</li>
                    <li><strong>Date:</strong> {{payment_date}}</li>
                    <li><strong>Method:</strong> {{payment_method}}</li>
                    <li><strong>Status:</strong> <span class="status-pending">{{payment_status}}</span></li>
                </ul>
            </div>

            <p>This payment is currently <strong>pending verification</strong>.</p>
            
            <p>Best regards,<br>Pay Your Dues Team</p>
        </div>
        <div class="footer">
            <p>This is an automated notification. Please do not reply to this email.</p>
        </div>
    </div>
</body>
</html>`
}

// GetPaymentVerifiedEmailTemplate returns a template for payment verification notifications
func GetPaymentVerifiedEmailTemplate() string {
	return `<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #4caf50; color: white; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
        .content { padding: 20px; background-color: #f9f9f9; }
        .details { background-color: white; padding: 15px; margin: 10px 0; border-left: 4px solid #4caf50; border-radius: 3px; }
        .footer { padding: 20px; text-align: center; font-size: 12px; color: #666; }
        .status-verified { color: #4caf50; font-weight: bold; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h2>✓ Payment Verified</h2>
        </div>
        <div class="content">
            <p>Hi {{recipient_name}},</p>
            <p>Great news! Your payment to <strong>{{other_party_name}}</strong> has been verified and accepted.</p>
            
            <div class="details">
                <h3>Payment Information:</h3>
                <ul>
                    <li><strong>Amount:</strong> {{currency}} {{payment_amount}}</li>
                    <li><strong>Date:</strong> {{payment_date}}</li>
                    <li><strong>Status:</strong> <span class="status-verified">Verified & Completed</span></li>
                </ul>
            </div>

            <p>Your payment has been successfully recorded.</p>
            
            <p>Best regards,<br>Pay Your Dues Team</p>
        </div>
        <div class="footer">
            <p>This is an automated notification. Please do not reply to this email.</p>
        </div>
    </div>
</body>
</html>`
}

// GetPaymentRejectedEmailTemplate returns a template for payment rejection notifications
func GetPaymentRejectedEmailTemplate() string {
	return `<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #f44336; color: white; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
        .content { padding: 20px; background-color: #f9f9f9; }
        .details { background-color: white; padding: 15px; margin: 10px 0; border-left: 4px solid #f44336; border-radius: 3px; }
        .footer { padding: 20px; text-align: center; font-size: 12px; color: #666; }
        .status-rejected { color: #f44336; font-weight: bold; }
        .reason { background-color: #ffebee; padding: 10px; border-radius: 3px; margin: 10px 0; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h2>✗ Payment Rejected</h2>
        </div>
        <div class="content">
            <p>Hi {{recipient_name}},</p>
            <p>Unfortunately, your payment to <strong>{{other_party_name}}</strong> has been rejected.</p>
            
            <div class="details">
                <h3>Payment Information:</h3>
                <ul>
                    <li><strong>Amount:</strong> {{currency}} {{payment_amount}}</li>
                    <li><strong>Date:</strong> {{payment_date}}</li>
                    <li><strong>Status:</strong> <span class="status-rejected">Rejected</span></li>
                </ul>
            </div>

            <div class="reason">
                <strong>Reason:</strong> {{rejection_reason}}
            </div>

            <p>Please contact <strong>{{other_party_name}}</strong> to resolve this issue and resubmit your payment.</p>
            
            <p>Best regards,<br>Pay Your Dues Team</p>
        </div>
        <div class="footer">
            <p>This is an automated notification. Please do not reply to this email.</p>
        </div>
    </div>
</body>
</html>`
}

