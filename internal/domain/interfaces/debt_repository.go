package interfaces

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"pay-your-dues/internal/domain/entities"
)

// DebtListRepository defines the interface for debt list data access operations
type DebtListRepository interface {
	Create(ctx context.Context, debtList *entities.DebtList) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.DebtList, error)
	GetByIDWithRelations(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*entities.DebtListResponse, error)
	GetUserDebtLists(ctx context.Context, userID uuid.UUID) ([]entities.DebtListResponse, error)
	GetDebtListsWhereUserIsContact(ctx context.Context, userID uuid.UUID) ([]entities.DebtListResponse, error)
	Update(ctx context.Context, debtList *entities.DebtList) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetOverdueForUser(ctx context.Context, userID uuid.UUID) ([]entities.DebtList, error)
	GetDueSoonForUser(ctx context.Context, userID uuid.UUID, dueDate time.Time) ([]entities.DebtList, error)
	BelongsToUser(ctx context.Context, debtListID, userID uuid.UUID) (bool, error)
	IsContactOfDebtList(ctx context.Context, debtListID, userID uuid.UUID) (bool, error)
	UpdatePaymentTotals(ctx context.Context, debtListID uuid.UUID, totalPaid, remaining decimal.Decimal) error
	UpdateStatus(ctx context.Context, debtListID uuid.UUID, status string) error
	UpdateNextPaymentDate(ctx context.Context, debtListID uuid.UUID, nextPaymentDate time.Time) error
	UpdateNotificationSettings(ctx context.Context, debtListID uuid.UUID, updates map[string]interface{}) error
}

// DebtItemRepository defines the interface for debt item data access operations
type DebtItemRepository interface {
	Create(ctx context.Context, debtItem *entities.DebtItem) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.DebtItem, error)
	GetByDebtListID(ctx context.Context, debtListID uuid.UUID) ([]entities.DebtItem, error)
	Update(ctx context.Context, debtItem *entities.DebtItem) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetTotalPaidForDebtList(ctx context.Context, debtListID uuid.UUID) (decimal.Decimal, error)
	GetCompletedPaymentsForDebtList(ctx context.Context, debtListID uuid.UUID) ([]entities.DebtItem, error)
	GetLastPaymentDate(ctx context.Context, debtListID uuid.UUID) (*time.Time, error)
	BelongsToUserDebtList(ctx context.Context, debtItemID, userID uuid.UUID) (bool, error)
	CanUserVerifyDebtItem(ctx context.Context, debtItemID, userID uuid.UUID) (bool, error)
	
	// Verification methods
	GetPendingVerifications(ctx context.Context, userID uuid.UUID) ([]entities.DebtItem, error)
	UpdatePaymentStatus(ctx context.Context, debtItemID uuid.UUID, status string, verifiedBy uuid.UUID, notes *string) error
	UpdateReceiptPhoto(ctx context.Context, debtItemID uuid.UUID, photoURL *string) error
}
