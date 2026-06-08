package mocks

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"

	"pay-your-dues/internal/domain/entities"
)

// MockDebtListRepository is a mock implementation of DebtListRepository
type MockDebtListRepository struct {
	mock.Mock
}

func (m *MockDebtListRepository) Create(ctx context.Context, debtList *entities.DebtList) error {
	args := m.Called(ctx, debtList)
	return args.Error(0)
}

func (m *MockDebtListRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.DebtList, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.DebtList), args.Error(1)
}

func (m *MockDebtListRepository) GetByIDWithRelations(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*entities.DebtListResponse, error) {
	args := m.Called(ctx, id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.DebtListResponse), args.Error(1)
}

func (m *MockDebtListRepository) GetUserDebtLists(ctx context.Context, userID uuid.UUID) ([]entities.DebtListResponse, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entities.DebtListResponse), args.Error(1)
}

func (m *MockDebtListRepository) Update(ctx context.Context, debtList *entities.DebtList) error {
	args := m.Called(ctx, debtList)
	return args.Error(0)
}

func (m *MockDebtListRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockDebtListRepository) BelongsToUser(ctx context.Context, debtListID uuid.UUID, userID uuid.UUID) (bool, error) {
	args := m.Called(ctx, debtListID, userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockDebtListRepository) IsContactOfDebtList(ctx context.Context, debtListID uuid.UUID, userID uuid.UUID) (bool, error) {
	args := m.Called(ctx, debtListID, userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockDebtListRepository) GetOverdueForUser(ctx context.Context, userID uuid.UUID) ([]entities.DebtList, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entities.DebtList), args.Error(1)
}

func (m *MockDebtListRepository) GetDueSoonForUser(ctx context.Context, userID uuid.UUID, dueDate time.Time) ([]entities.DebtList, error) {
	args := m.Called(ctx, userID, dueDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entities.DebtList), args.Error(1)
}

func (m *MockDebtListRepository) UpdatePaymentTotals(ctx context.Context, debtListID uuid.UUID, totalPaid, remainingDebt decimal.Decimal) error {
	args := m.Called(ctx, debtListID, totalPaid, remainingDebt)
	return args.Error(0)
}

func (m *MockDebtListRepository) UpdateStatus(ctx context.Context, debtListID uuid.UUID, status string) error {
	args := m.Called(ctx, debtListID, status)
	return args.Error(0)
}

func (m *MockDebtListRepository) UpdateNextPaymentDate(ctx context.Context, debtListID uuid.UUID, nextPaymentDate time.Time) error {
	args := m.Called(ctx, debtListID, nextPaymentDate)
	return args.Error(0)
}

func (m *MockDebtListRepository) UpdateNotificationSettings(ctx context.Context, debtListID uuid.UUID, updates map[string]interface{}) error {
	args := m.Called(ctx, debtListID, updates)
	return args.Error(0)
}

func (m *MockDebtListRepository) GetDebtListsWhereUserIsContact(ctx context.Context, userID uuid.UUID) ([]entities.DebtListResponse, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entities.DebtListResponse), args.Error(1)
}

// MockDebtItemRepository is a mock implementation of DebtItemRepository
type MockDebtItemRepository struct {
	mock.Mock
}

func (m *MockDebtItemRepository) Create(ctx context.Context, debtItem *entities.DebtItem) error {
	args := m.Called(ctx, debtItem)
	return args.Error(0)
}

func (m *MockDebtItemRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.DebtItem, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.DebtItem), args.Error(1)
}

func (m *MockDebtItemRepository) GetByDebtListID(ctx context.Context, debtListID uuid.UUID) ([]entities.DebtItem, error) {
	args := m.Called(ctx, debtListID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entities.DebtItem), args.Error(1)
}

func (m *MockDebtItemRepository) Update(ctx context.Context, debtItem *entities.DebtItem) error {
	args := m.Called(ctx, debtItem)
	return args.Error(0)
}

func (m *MockDebtItemRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockDebtItemRepository) BelongsToUserDebtList(ctx context.Context, debtItemID uuid.UUID, userID uuid.UUID) (bool, error) {
	args := m.Called(ctx, debtItemID, userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockDebtItemRepository) CanUserVerifyDebtItem(ctx context.Context, debtItemID uuid.UUID, userID uuid.UUID) (bool, error) {
	args := m.Called(ctx, debtItemID, userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockDebtItemRepository) GetCompletedPaymentsForDebtList(ctx context.Context, debtListID uuid.UUID) ([]entities.DebtItem, error) {
	args := m.Called(ctx, debtListID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entities.DebtItem), args.Error(1)
}

func (m *MockDebtItemRepository) GetTotalPaidForDebtList(ctx context.Context, debtListID uuid.UUID) (decimal.Decimal, error) {
	args := m.Called(ctx, debtListID)
	return args.Get(0).(decimal.Decimal), args.Error(1)
}

func (m *MockDebtItemRepository) GetLastPaymentDate(ctx context.Context, debtListID uuid.UUID) (*time.Time, error) {
	args := m.Called(ctx, debtListID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*time.Time), args.Error(1)
}

// Verification methods
func (m *MockDebtItemRepository) GetPendingVerifications(ctx context.Context, userID uuid.UUID) ([]entities.DebtItem, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entities.DebtItem), args.Error(1)
}

func (m *MockDebtItemRepository) UpdatePaymentStatus(ctx context.Context, debtItemID uuid.UUID, status string, verifiedBy uuid.UUID, notes *string) error {
	args := m.Called(ctx, debtItemID, status, verifiedBy, notes)
	return args.Error(0)
}

func (m *MockDebtItemRepository) UpdateReceiptPhoto(ctx context.Context, debtItemID uuid.UUID, photoURL *string) error {
	args := m.Called(ctx, debtItemID, photoURL)
	return args.Error(0)
}

// MockPaymentScheduleService is a mock implementation of PaymentScheduleService
type MockPaymentScheduleService struct {
	mock.Mock
}

func (m *MockPaymentScheduleService) CalculateNextPaymentDate(debtList *entities.DebtList, lastPaymentDate *time.Time) time.Time {
	args := m.Called(debtList, lastPaymentDate)
	return args.Get(0).(time.Time)
}

func (m *MockPaymentScheduleService) CalculatePaymentSchedule(debtList *entities.DebtList, payments []entities.DebtItem) []entities.PaymentScheduleItem {
	args := m.Called(debtList, payments)
	return args.Get(0).([]entities.PaymentScheduleItem)
}

func (m *MockPaymentScheduleService) CalculateDueDateFromNumberOfPayments(createdAt time.Time, numberOfPayments int, installmentPlan string) time.Time {
	args := m.Called(createdAt, numberOfPayments, installmentPlan)
	return args.Get(0).(time.Time)
}

func (m *MockPaymentScheduleService) CalculateInstallmentAmountFromNumberOfPayments(totalAmount decimal.Decimal, numberOfPayments int) decimal.Decimal {
	args := m.Called(totalAmount, numberOfPayments)
	return args.Get(0).(decimal.Decimal)
}

func (m *MockPaymentScheduleService) CalculateInstallmentAmount(totalAmount decimal.Decimal, installmentPlan string, createdAt time.Time, dueDate time.Time) decimal.Decimal {
	args := m.Called(totalAmount, installmentPlan, createdAt, dueDate)
	return args.Get(0).(decimal.Decimal)
}
