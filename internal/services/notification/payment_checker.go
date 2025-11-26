package notification

import (
	"time"

	"pay-your-dues/internal/domain/interfaces"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// PaymentChecker handles checking payment status for notifications
type PaymentChecker struct {
	db           *gorm.DB
	debtListRepo interfaces.DebtListRepository
	debtItemRepo interfaces.DebtItemRepository
}

// NewPaymentChecker creates a new payment checker instance
func NewPaymentChecker(
	db *gorm.DB,
	debtListRepo interfaces.DebtListRepository,
	debtItemRepo interfaces.DebtItemRepository,
) *PaymentChecker {
	return &PaymentChecker{
		db:           db,
		debtListRepo: debtListRepo,
		debtItemRepo: debtItemRepo,
	}
}

// InstallmentStatus represents the status of a specific installment
type InstallmentStatus struct {
	PaymentNumber    int
	DueDate          time.Time
	ScheduledAmount  decimal.Decimal
	PaidAmount       decimal.Decimal
	Status           string // "pending", "partial", "paid"
	RemainingAmount  decimal.Decimal
}

// ShouldSendNotificationForInstallment checks if a notification should be sent for a specific installment
func (pc *PaymentChecker) ShouldSendNotificationForInstallment(debtListID uuid.UUID, installmentNumber *int) (bool, error) {
	// If no installment number, check overall debt status (one-time debts)
	if installmentNumber == nil {
		return pc.shouldSendForOneTimeDebt(debtListID)
	}

	// Get installment status
	status, err := pc.GetInstallmentStatus(debtListID, *installmentNumber)
	if err != nil {
		return false, err
	}

	// Don't send if installment is paid
	if status.Status == "paid" {
		return false, nil
	}

	// Don't send if paid amount >= scheduled amount
	if status.PaidAmount.GreaterThanOrEqual(status.ScheduledAmount) {
		return false, nil
	}

	// Check if entire debt is settled
	isSettled, err := pc.IsDebtSettled(debtListID)
	if err != nil {
		return false, err
	}

	if isSettled {
		return false, nil
	}

	return true, nil
}

// GetInstallmentStatus gets the status of a specific installment
func (pc *PaymentChecker) GetInstallmentStatus(debtListID uuid.UUID, installmentNumber int) (*InstallmentStatus, error) {
	var status InstallmentStatus

	query := `
		WITH payment_schedule AS (
			SELECT
				ROW_NUMBER() OVER (ORDER BY created_at) as payment_number,
				due_date,
				installment_amount as scheduled_amount
			FROM generate_series(
				(SELECT due_date FROM debt_lists WHERE id = $1),
				(SELECT due_date + (number_of_payments - 1) * 
					CASE installment_plan
						WHEN 'weekly' THEN interval '7 days'
						WHEN 'biweekly' THEN interval '14 days'
						WHEN 'monthly' THEN interval '1 month'
						WHEN 'quarterly' THEN interval '3 months'
						WHEN 'yearly' THEN interval '1 year'
					END
				FROM debt_lists WHERE id = $1),
				CASE (SELECT installment_plan FROM debt_lists WHERE id = $1)
					WHEN 'weekly' THEN interval '7 days'
					WHEN 'biweekly' THEN interval '14 days'
					WHEN 'monthly' THEN interval '1 month'
					WHEN 'quarterly' THEN interval '3 months'
					WHEN 'yearly' THEN interval '1 year'
				END
			) as due_date
		),
		payments_made AS (
			SELECT
				COALESCE(SUM(amount), 0) as paid_amount
			FROM debt_items
			WHERE debt_list_id = $1
				AND status = 'completed'
				AND payment_date <= (
					SELECT due_date FROM payment_schedule WHERE payment_number = $2
				)
		)
		SELECT
			ps.payment_number,
			ps.due_date,
			ps.scheduled_amount,
			COALESCE(pm.paid_amount, 0) as paid_amount,
			CASE
				WHEN COALESCE(pm.paid_amount, 0) >= ps.scheduled_amount THEN 'paid'
				WHEN COALESCE(pm.paid_amount, 0) > 0 THEN 'partial'
				ELSE 'pending'
			END as status,
			GREATEST(ps.scheduled_amount - COALESCE(pm.paid_amount, 0), 0) as remaining_amount
		FROM payment_schedule ps
		CROSS JOIN payments_made pm
		WHERE ps.payment_number = $2
	`

	err := pc.db.Raw(query, debtListID, installmentNumber).Scan(&status).Error
	if err != nil {
		return nil, err
	}

	return &status, nil
}

// shouldSendForOneTimeDebt checks if notification should be sent for one-time debts
func (pc *PaymentChecker) shouldSendForOneTimeDebt(debtListID uuid.UUID) (bool, error) {
	var result struct {
		Status        string
		RemainingDebt decimal.Decimal
	}

	query := `
		SELECT
			status,
			total_remaining_debt as remaining_debt
		FROM debt_lists
		WHERE id = ?
	`

	err := pc.db.Raw(query, debtListID).Scan(&result).Error
	if err != nil {
		return false, err
	}

	// Don't send if debt is settled
	if result.Status == "settled" {
		return false, nil
	}

	// Don't send if remaining debt is zero or negative
	if result.RemainingDebt.LessThanOrEqual(decimal.Zero) {
		return false, nil
	}

	return true, nil
}

// IsDebtSettled checks if the entire debt is settled
func (pc *PaymentChecker) IsDebtSettled(debtListID uuid.UUID) (bool, error) {
	var result struct {
		Status        string
		RemainingDebt decimal.Decimal
	}

	query := `
		SELECT
			status,
			total_remaining_debt as remaining_debt
		FROM debt_lists
		WHERE id = ?
	`

	err := pc.db.Raw(query, debtListID).Scan(&result).Error
	if err != nil {
		return false, err
	}

	return result.Status == "settled" || result.RemainingDebt.LessThanOrEqual(decimal.Zero), nil
}

// GetPaymentProgress returns the payment progress for a debt
func (pc *PaymentChecker) GetPaymentProgress(debtListID uuid.UUID) (*PaymentProgress, error) {
	var progress PaymentProgress

	query := `
		SELECT
			dl.id as debt_list_id,
			dl.total_debt,
			dl.total_remaining_debt as remaining_debt,
			dl.total_debt - dl.total_remaining_debt as paid_amount,
			CASE
				WHEN dl.total_debt > 0 THEN
					ROUND(((dl.total_debt - dl.total_remaining_debt) / dl.total_debt * 100)::numeric, 2)
				ELSE 0
			END as progress_percentage,
			dl.status,
			COALESCE(dl.number_of_payments, 1) as total_installments,
			COALESCE(
				(SELECT COUNT(*) 
				 FROM debt_items di 
				 WHERE di.debt_list_id = dl.id 
				   AND di.status = 'completed'
				), 0
			) as paid_installments
		FROM debt_lists dl
		WHERE dl.id = ?
	`

	err := pc.db.Raw(query, debtListID).Scan(&progress).Error
	if err != nil {
		return nil, err
	}

	progress.RemainingInstallments = progress.TotalInstallments - progress.PaidInstallments

	return &progress, nil
}

// PaymentProgress represents the overall payment progress
type PaymentProgress struct {
	DebtListID            uuid.UUID
	TotalDebt             decimal.Decimal
	RemainingDebt         decimal.Decimal
	PaidAmount            decimal.Decimal
	ProgressPercentage    float64
	Status                string
	TotalInstallments     int
	PaidInstallments      int
	RemainingInstallments int
}

// IsOverdue checks if a debt or installment is overdue
func (pc *PaymentChecker) IsOverdue(debtListID uuid.UUID, installmentNumber *int) (bool, error) {
	var dueDate time.Time

	if installmentNumber == nil {
		// Check one-time debt due date
		err := pc.db.Raw(`
			SELECT due_date
			FROM debt_lists
			WHERE id = ?
		`, debtListID).Scan(&dueDate).Error
		if err != nil {
			return false, err
		}
	} else {
		// Check specific installment due date
		status, err := pc.GetInstallmentStatus(debtListID, *installmentNumber)
		if err != nil {
			return false, err
		}
		dueDate = status.DueDate
	}

	return time.Now().After(dueDate), nil
}

// GetUpcomingDueDate returns the next upcoming due date for a debt
func (pc *PaymentChecker) GetUpcomingDueDate(debtListID uuid.UUID) (*time.Time, error) {
	var dueDate *time.Time

	query := `
		SELECT
			CASE
				WHEN dl.debt_type = 'onetime' THEN dl.due_date
				ELSE dl.next_payment_date
			END as due_date
		FROM debt_lists dl
		WHERE dl.id = ?
			AND dl.status != 'settled'
			AND dl.total_remaining_debt > 0
	`

	err := pc.db.Raw(query, debtListID).Scan(&dueDate).Error
	if err != nil {
		return nil, err
	}

	return dueDate, nil
}

