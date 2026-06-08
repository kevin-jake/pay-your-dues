package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"
)

// Struct-level validation for CreateDebtListRequest
func (req CreateDebtListRequest) Validate() error {
	hasDueDate := req.DueDate != nil
	hasNumberOfPayments := req.NumberOfPayments != nil && *req.NumberOfPayments > 0
	
	if !hasDueDate && !hasNumberOfPayments {
		return fmt.Errorf("either due_date or number_of_payments must be provided")
	}
	
	// If number_of_payments is provided, installment_plan is required
	if hasNumberOfPayments && req.InstallmentPlan == "" {
		return fmt.Errorf("installment_plan is required when number_of_payments is provided")
	}
	
	return nil
}

type DebtList struct {
	ID              uuid.UUID     `json:"id" gorm:"type:uuid;primary_key"`
	UserID          uuid.UUID     `json:"user_id" gorm:"type:uuid;not null;index"`
	ContactID       uuid.UUID     `json:"contact_id" gorm:"type:uuid;not null;index"`
	DebtType        string        `json:"debt_type" gorm:"not null;index;check:debt_type IN ('to_receive', 'to_pay')"`
	TotalAmount     decimal.Decimal `json:"total_amount" gorm:"type:decimal(15,2);not null"`
	InstallmentAmount decimal.Decimal `json:"installment_amount" gorm:"type:decimal(15,2);not null"`
	TotalPaymentsMade decimal.Decimal `json:"total_payments_made" gorm:"type:decimal(15,2);default:0"`
	TotalRemainingDebt decimal.Decimal `json:"total_remaining_debt" gorm:"type:decimal(15,2);not null"`
	Currency        string        `json:"currency" gorm:"default:'Php'"`
	Status          string        `json:"status" gorm:"default:'active';index;check:status IN ('active', 'settled', 'archived', 'overdue')"`
	DueDate         time.Time     `json:"due_date" gorm:"not null"`
	NextPaymentDate time.Time     `json:"next_payment_date" gorm:"not null"`
	InstallmentPlan string        `json:"installment_plan" gorm:"default:'monthly';check:installment_plan IN ('onetime', 'weekly', 'biweekly', 'monthly', 'quarterly', 'yearly')"`
	NumberOfPayments *int         `json:"number_of_payments" gorm:"default:null"`
	Description     *string       `json:"description"`
	Notes           *string       `json:"notes"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`

	// Per-debt notification settings
	NotificationsEnabled         bool          `json:"notifications_enabled" gorm:"default:true"`
	NotificationReminderDays     pq.Int64Array `json:"notification_reminder_days" gorm:"type:integer[]"`
	NotificationTime             *string       `json:"notification_time"`
	NotifyEmail                  *bool         `json:"notify_email"`
	NotifySMS                    *bool         `json:"notify_sms"`
	NotifySlack                  *bool         `json:"notify_slack"`
	NotifyTelegram               *bool         `json:"notify_telegram"`
	NotifyDiscord                *bool         `json:"notify_discord"`

	// Relationships
	User      User        `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Contact   Contact     `json:"contact,omitempty" gorm:"foreignKey:ContactID"`
	Payments  []DebtItem  `json:"payments,omitempty" gorm:"foreignKey:DebtListID"`
}

type CreateDebtListRequest struct {
	ContactID         uuid.UUID `json:"contact_id" binding:"required"`
	DebtType          string    `json:"debt_type" binding:"required,oneof=to_receive to_pay"`
	TotalAmount       string    `json:"total_amount" binding:"required"`
	Currency          string    `json:"currency"`
	DueDate           *time.Time `json:"due_date"`
	InstallmentPlan   string    `json:"installment_plan" binding:"omitempty,oneof=onetime weekly biweekly monthly quarterly yearly"`
	NumberOfPayments  *int      `json:"number_of_payments"`
	Description       *string   `json:"description"`
	Notes             *string   `json:"notes"`
}

type UpdateDebtListRequest struct {
	TotalAmount       *string    `json:"total_amount"`
	Currency          *string    `json:"currency"`
	Status            *string    `json:"status"`
	DueDate           *time.Time `json:"due_date"`
	InstallmentPlan   *string    `json:"installment_plan"`
	NumberOfPayments  *int       `json:"number_of_payments"`
	Description       *string    `json:"description"`
	Notes             *string    `json:"notes"`
}

// DebtItemResponse represents a payment without the circular reference to DebtList
type DebtItemResponse struct {
	ID            uuid.UUID     `json:"id"`
	DebtListID    uuid.UUID     `json:"debt_list_id"`
	Amount        decimal.Decimal `json:"amount"`
	Currency      string        `json:"currency"`
	PaymentDate   time.Time     `json:"payment_date"`
	PaymentMethod string        `json:"payment_method"`
	Description   *string       `json:"description"`
	Status        string        `json:"status"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

// DebtListResponse represents a debt list with payments but without circular references
type DebtListResponse struct {
	ID              uuid.UUID     `json:"id"`
	UserID          uuid.UUID     `json:"user_id"`
	ContactID       uuid.UUID     `json:"contact_id"`
	DebtType        string        `json:"debt_type"`
	TotalAmount     decimal.Decimal `json:"total_amount"`
	InstallmentAmount decimal.Decimal `json:"installment_amount"`
	TotalPaymentsMade decimal.Decimal `json:"total_payments_made"`
	TotalRemainingDebt decimal.Decimal `json:"total_remaining_debt"`
	Currency        string        `json:"currency"`
	Status          string        `json:"status"`
	DueDate         time.Time     `json:"due_date"`
	NextPaymentDate time.Time     `json:"next_payment_date"`
	InstallmentPlan string        `json:"installment_plan"`
	NumberOfPayments *int         `json:"number_of_payments"`
	Description     *string       `json:"description"`
	Notes           *string       `json:"notes"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
	
	// Relationships
	User      User              `json:"user,omitempty"`
	Contact   Contact           `json:"contact,omitempty"`
	Payments  []DebtItemResponse `json:"payments,omitempty"`
}

// ToDebtItemResponse converts a DebtItem to DebtItemResponse
func (di *DebtItem) ToDebtItemResponse() DebtItemResponse {
	return DebtItemResponse{
		ID:            di.ID,
		DebtListID:    di.DebtListID,
		Amount:        di.Amount,
		Currency:      di.Currency,
		PaymentDate:   di.PaymentDate,
		PaymentMethod: di.PaymentMethod,
		Description:   di.Description,
		Status:        di.Status,
		CreatedAt:     di.CreatedAt,
		UpdatedAt:     di.UpdatedAt,
	}
}

// ToDebtListResponse converts a DebtList to DebtListResponse
func (dl *DebtList) ToDebtListResponse() DebtListResponse {
	payments := make([]DebtItemResponse, len(dl.Payments))
	for i, payment := range dl.Payments {
		payments[i] = payment.ToDebtItemResponse()
	}

	return DebtListResponse{
		ID:              dl.ID,
		UserID:          dl.UserID,
		ContactID:       dl.ContactID,
		DebtType:        dl.DebtType,
		TotalAmount:     dl.TotalAmount,
		InstallmentAmount: dl.InstallmentAmount,
		TotalPaymentsMade: dl.TotalPaymentsMade,
		TotalRemainingDebt: dl.TotalRemainingDebt,
		Currency:        dl.Currency,
		Status:          dl.Status,
		DueDate:         dl.DueDate,
		NextPaymentDate: dl.NextPaymentDate,
		InstallmentPlan: dl.InstallmentPlan,
		NumberOfPayments: dl.NumberOfPayments,
		Description:     dl.Description,
		Notes:           dl.Notes,
		CreatedAt:       dl.CreatedAt,
		UpdatedAt:       dl.UpdatedAt,
		User:            dl.User,
		Contact:         dl.Contact,
		Payments:        payments,
	}
} 