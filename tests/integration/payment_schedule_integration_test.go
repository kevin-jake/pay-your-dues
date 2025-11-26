package integration

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"pay-your-dues/internal/domain/entities"
	"pay-your-dues/internal/domain/interfaces"
	"pay-your-dues/internal/mocks"
	"pay-your-dues/internal/services"
)

type PaymentScheduleIntegrationTestSuite struct {
	suite.Suite
	debtService            interfaces.DebtService
	paymentScheduleService interfaces.PaymentScheduleService
	debtListRepo           *mocks.MockDebtListRepository
	debtItemRepo           *mocks.MockDebtItemRepository
	contactRepo            *mocks.MockContactRepository
	fileStorageService     *mocks.MockFileStorageService
}

func (suite *PaymentScheduleIntegrationTestSuite) SetupTest() {
	suite.debtListRepo = new(mocks.MockDebtListRepository)
	suite.debtItemRepo = new(mocks.MockDebtItemRepository)
	suite.contactRepo = new(mocks.MockContactRepository)
	suite.fileStorageService = new(mocks.MockFileStorageService)
	suite.paymentScheduleService = services.NewPaymentScheduleService()
	
	suite.debtService = services.NewDebtService(
		suite.debtListRepo,
		suite.debtItemRepo,
		suite.contactRepo,
		suite.paymentScheduleService,
		suite.fileStorageService,
		nil,
	)
}

func TestPaymentScheduleIntegrationTestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	suite.Run(t, new(PaymentScheduleIntegrationTestSuite))
}

func (suite *PaymentScheduleIntegrationTestSuite) TestGetPaymentSchedule_NoPayments() {
	t := suite.T()
	ctx := context.Background()

	userID := uuid.New()
	debtListID := uuid.New()
	createdAt := time.Now().AddDate(0, 0, -30) // Created 30 days ago

	debtList := &entities.DebtList{
		ID:                debtListID,
		UserID:            userID,
		TotalAmount:       decimal.RequireFromString("1000.00"),
		InstallmentAmount: decimal.RequireFromString("250.00"),
		InstallmentPlan:   "monthly",
		CreatedAt:         createdAt,
		DueDate:           createdAt.AddDate(0, 4, 0),
	}

	// Mock expectations
	suite.debtListRepo.On("BelongsToUser", ctx, debtListID, userID).Return(true, nil)
	suite.debtListRepo.On("GetByID", ctx, debtListID).Return(debtList, nil)
	suite.debtItemRepo.On("GetCompletedPaymentsForDebtList", ctx, debtListID).Return([]entities.DebtItem{}, nil)

	// Execute
	schedule, err := suite.debtService.GetPaymentSchedule(ctx, debtListID, userID)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, schedule)
	assert.Len(t, schedule, 4, "Should have 4 monthly payments")

	// Verify all payments are pending with correct amounts
	for i, item := range schedule {
		assert.Equal(t, i+1, item.PaymentNumber)
		assert.Equal(t, "pending", item.Status)
		assert.True(t, item.PaidAmount.IsZero(), "No payments made yet")
		assert.True(t, item.ScheduledAmount.Equal(decimal.RequireFromString("250.00")))
		assert.True(t, item.Amount.Equal(decimal.RequireFromString("250.00")), "Full amount still owed")
	}

	suite.debtListRepo.AssertExpectations(t)
	suite.debtItemRepo.AssertExpectations(t)
}

func (suite *PaymentScheduleIntegrationTestSuite) TestGetPaymentSchedule_WithFullPayments() {
	t := suite.T()
	ctx := context.Background()

	userID := uuid.New()
	debtListID := uuid.New()
	createdAt := time.Now().AddDate(0, 0, -60) // Created 60 days ago

	debtList := &entities.DebtList{
		ID:                debtListID,
		UserID:            userID,
		TotalAmount:       decimal.RequireFromString("1000.00"),
		InstallmentAmount: decimal.RequireFromString("250.00"),
		InstallmentPlan:   "monthly",
		CreatedAt:         createdAt,
		DueDate:           createdAt.AddDate(0, 4, 0),
	}

	payments := []entities.DebtItem{
		{
			ID:          uuid.New(),
			DebtListID:  debtListID,
			Amount:      decimal.RequireFromString("250.00"),
			Status:      "completed",
			PaymentDate: createdAt.AddDate(0, 0, 10),
		},
		{
			ID:          uuid.New(),
			DebtListID:  debtListID,
			Amount:      decimal.RequireFromString("250.00"),
			Status:      "completed",
			PaymentDate: createdAt.AddDate(0, 1, 5),
		},
	}

	// Mock expectations
	suite.debtListRepo.On("BelongsToUser", ctx, debtListID, userID).Return(true, nil)
	suite.debtListRepo.On("GetByID", ctx, debtListID).Return(debtList, nil)
	suite.debtItemRepo.On("GetCompletedPaymentsForDebtList", ctx, debtListID).Return(payments, nil)

	// Execute
	schedule, err := suite.debtService.GetPaymentSchedule(ctx, debtListID, userID)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, schedule)
	assert.Len(t, schedule, 4)

	// First two payments should be paid
	for i := 0; i < 2; i++ {
		assert.Equal(t, "paid", schedule[i].Status)
		assert.True(t, schedule[i].PaidAmount.Equal(decimal.RequireFromString("250.00")))
		assert.True(t, schedule[i].ScheduledAmount.Equal(decimal.RequireFromString("250.00")))
		assert.True(t, schedule[i].Amount.IsZero(), "Should be fully paid")
	}

	// Last two payments should be pending
	for i := 2; i < 4; i++ {
		assert.Equal(t, "pending", schedule[i].Status)
		assert.True(t, schedule[i].PaidAmount.IsZero())
		assert.True(t, schedule[i].Amount.Equal(decimal.RequireFromString("250.00")))
	}

	suite.debtListRepo.AssertExpectations(t)
	suite.debtItemRepo.AssertExpectations(t)
}

func (suite *PaymentScheduleIntegrationTestSuite) TestGetPaymentSchedule_WithPartialPayment() {
	t := suite.T()
	ctx := context.Background()

	userID := uuid.New()
	debtListID := uuid.New()
	createdAt := time.Now().AddDate(0, 0, -45)

	debtList := &entities.DebtList{
		ID:                debtListID,
		UserID:            userID,
		TotalAmount:       decimal.RequireFromString("1000.00"),
		InstallmentAmount: decimal.RequireFromString("250.00"),
		InstallmentPlan:   "monthly",
		CreatedAt:         createdAt,
		DueDate:           createdAt.AddDate(0, 4, 0),
	}

	payments := []entities.DebtItem{
		{
			ID:          uuid.New(),
			DebtListID:  debtListID,
			Amount:      decimal.RequireFromString("250.00"),
			Status:      "completed",
			PaymentDate: createdAt.AddDate(0, 0, 10),
		},
		{
			ID:          uuid.New(),
			DebtListID:  debtListID,
			Amount:      decimal.RequireFromString("150.00"), // Partial payment
			Status:      "completed",
			PaymentDate: createdAt.AddDate(0, 1, 5),
		},
	}

	// Mock expectations
	suite.debtListRepo.On("BelongsToUser", ctx, debtListID, userID).Return(true, nil)
	suite.debtListRepo.On("GetByID", ctx, debtListID).Return(debtList, nil)
	suite.debtItemRepo.On("GetCompletedPaymentsForDebtList", ctx, debtListID).Return(payments, nil)

	// Execute
	schedule, err := suite.debtService.GetPaymentSchedule(ctx, debtListID, userID)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, schedule)
	assert.Len(t, schedule, 4)

	// First payment should be fully paid
	assert.Equal(t, "paid", schedule[0].Status)
	assert.True(t, schedule[0].PaidAmount.Equal(decimal.RequireFromString("250.00")))
	assert.True(t, schedule[0].ScheduledAmount.Equal(decimal.RequireFromString("250.00")))
	assert.True(t, schedule[0].Amount.IsZero())

	// Second payment should be partially paid
	assert.Equal(t, "pending", schedule[1].Status, "Partial payment should still be pending")
	assert.True(t, schedule[1].PaidAmount.Equal(decimal.RequireFromString("150.00")), "Should show 150 paid")
	assert.True(t, schedule[1].ScheduledAmount.Equal(decimal.RequireFromString("250.00")))
	assert.True(t, schedule[1].Amount.Equal(decimal.RequireFromString("100.00")), "Should still owe 100")

	// Remaining payments should be pending
	for i := 2; i < 4; i++ {
		assert.Equal(t, "pending", schedule[i].Status)
		assert.True(t, schedule[i].PaidAmount.IsZero())
		assert.True(t, schedule[i].Amount.Equal(decimal.RequireFromString("250.00")))
	}

	suite.debtListRepo.AssertExpectations(t)
	suite.debtItemRepo.AssertExpectations(t)
}

func (suite *PaymentScheduleIntegrationTestSuite) TestGetPaymentSchedule_AllPaid() {
	t := suite.T()
	ctx := context.Background()

	userID := uuid.New()
	debtListID := uuid.New()
	createdAt := time.Now().AddDate(0, -4, 0) // Created 4 months ago

	debtList := &entities.DebtList{
		ID:                debtListID,
		UserID:            userID,
		TotalAmount:       decimal.RequireFromString("1000.00"),
		InstallmentAmount: decimal.RequireFromString("250.00"),
		InstallmentPlan:   "monthly",
		CreatedAt:         createdAt,
		DueDate:           createdAt.AddDate(0, 4, 0),
	}

	// One large payment covering everything
	payments := []entities.DebtItem{
		{
			ID:          uuid.New(),
			DebtListID:  debtListID,
			Amount:      decimal.RequireFromString("1000.00"),
			Status:      "completed",
			PaymentDate: createdAt.AddDate(0, 2, 0),
		},
	}

	// Mock expectations
	suite.debtListRepo.On("BelongsToUser", ctx, debtListID, userID).Return(true, nil)
	suite.debtListRepo.On("GetByID", ctx, debtListID).Return(debtList, nil)
	suite.debtItemRepo.On("GetCompletedPaymentsForDebtList", ctx, debtListID).Return(payments, nil)

	// Execute
	schedule, err := suite.debtService.GetPaymentSchedule(ctx, debtListID, userID)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, schedule)
	assert.Len(t, schedule, 4)

	// All payments should be marked as paid
	for _, item := range schedule {
		assert.Equal(t, "paid", item.Status)
		assert.True(t, item.PaidAmount.Equal(decimal.RequireFromString("250.00")))
		assert.True(t, item.ScheduledAmount.Equal(decimal.RequireFromString("250.00")))
		assert.True(t, item.Amount.IsZero(), "All payments fully covered")
	}

	suite.debtListRepo.AssertExpectations(t)
	suite.debtItemRepo.AssertExpectations(t)
}

func (suite *PaymentScheduleIntegrationTestSuite) TestGetPaymentSchedule_WeeklyPlan() {
	t := suite.T()
	ctx := context.Background()

	userID := uuid.New()
	debtListID := uuid.New()
	createdAt := time.Now().AddDate(0, 0, -14)

	debtList := &entities.DebtList{
		ID:                debtListID,
		UserID:            userID,
		TotalAmount:       decimal.RequireFromString("400.00"),
		InstallmentAmount: decimal.RequireFromString("100.00"),
		InstallmentPlan:   "weekly",
		CreatedAt:         createdAt,
		DueDate:           createdAt.AddDate(0, 0, 28),
	}

	payments := []entities.DebtItem{
		{
			ID:          uuid.New(),
			DebtListID:  debtListID,
			Amount:      decimal.RequireFromString("100.00"),
			Status:      "completed",
			PaymentDate: createdAt.AddDate(0, 0, 3),
		},
	}

	// Mock expectations
	suite.debtListRepo.On("BelongsToUser", ctx, debtListID, userID).Return(true, nil)
	suite.debtListRepo.On("GetByID", ctx, debtListID).Return(debtList, nil)
	suite.debtItemRepo.On("GetCompletedPaymentsForDebtList", ctx, debtListID).Return(payments, nil)

	// Execute
	schedule, err := suite.debtService.GetPaymentSchedule(ctx, debtListID, userID)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, schedule)
	assert.Len(t, schedule, 4, "Should have 4 weekly payments")

	// First payment should be paid
	assert.Equal(t, "paid", schedule[0].Status)
	assert.True(t, schedule[0].PaidAmount.Equal(decimal.RequireFromString("100.00")))

	// Verify weekly intervals
	for i := 1; i < len(schedule); i++ {
		diff := schedule[i].DueDate.Sub(schedule[i-1].DueDate)
		days := int(diff.Hours() / 24)
		assert.Equal(t, 7, days, "Weekly payments should be 7 days apart")
	}

	suite.debtListRepo.AssertExpectations(t)
	suite.debtItemRepo.AssertExpectations(t)
}

func (suite *PaymentScheduleIntegrationTestSuite) TestGetPaymentSchedule_Unauthorized() {
	t := suite.T()
	ctx := context.Background()

	userID := uuid.New()
	debtListID := uuid.New()

	// Mock expectations - user doesn't own the debt and is not a contact
	suite.debtListRepo.On("BelongsToUser", ctx, debtListID, userID).Return(false, nil)
	suite.debtListRepo.On("IsContactOfDebtList", ctx, debtListID, userID).Return(false, nil)

	// Execute
	schedule, err := suite.debtService.GetPaymentSchedule(ctx, debtListID, userID)

	// Assert
	require.Error(t, err)
	assert.Nil(t, schedule)
	assert.Equal(t, entities.ErrDebtListNotFound, err)

	suite.debtListRepo.AssertExpectations(t)
}

func (suite *PaymentScheduleIntegrationTestSuite) TestGetPaymentSchedule_ContactCanView() {
	t := suite.T()
	ctx := context.Background()

	userID := uuid.New()
	debtListID := uuid.New()
	createdAt := time.Now().AddDate(0, 0, -30)

	debtList := &entities.DebtList{
		ID:                debtListID,
		UserID:            uuid.New(), // Different user owns it
		TotalAmount:       decimal.RequireFromString("500.00"),
		InstallmentAmount: decimal.RequireFromString("100.00"),
		InstallmentPlan:   "monthly",
		CreatedAt:         createdAt,
		DueDate:           createdAt.AddDate(0, 5, 0),
	}

	// Mock expectations - user is a contact of this debt
	suite.debtListRepo.On("BelongsToUser", ctx, debtListID, userID).Return(false, nil)
	suite.debtListRepo.On("IsContactOfDebtList", ctx, debtListID, userID).Return(true, nil)
	suite.debtListRepo.On("GetByID", ctx, debtListID).Return(debtList, nil)
	suite.debtItemRepo.On("GetCompletedPaymentsForDebtList", ctx, debtListID).Return([]entities.DebtItem{}, nil)

	// Execute
	schedule, err := suite.debtService.GetPaymentSchedule(ctx, debtListID, userID)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, schedule)
	assert.Len(t, schedule, 5, "Should have 5 monthly payments")

	suite.debtListRepo.AssertExpectations(t)
	suite.debtItemRepo.AssertExpectations(t)
}

func (suite *PaymentScheduleIntegrationTestSuite) TestGetPaymentSchedule_PendingPaymentsNotCounted() {
	t := suite.T()
	ctx := context.Background()

	userID := uuid.New()
	debtListID := uuid.New()
	createdAt := time.Now().AddDate(0, 0, -30)

	debtList := &entities.DebtList{
		ID:                debtListID,
		UserID:            userID,
		TotalAmount:       decimal.RequireFromString("1000.00"),
		InstallmentAmount: decimal.RequireFromString("250.00"),
		InstallmentPlan:   "monthly",
		CreatedAt:         createdAt,
		DueDate:           createdAt.AddDate(0, 4, 0),
	}

	// Only completed payments should be counted
	payments := []entities.DebtItem{
		{
			ID:          uuid.New(),
			DebtListID:  debtListID,
			Amount:      decimal.RequireFromString("250.00"),
			Status:      "completed",
			PaymentDate: createdAt.AddDate(0, 0, 10),
		},
		// Pending and other statuses should not be counted
	}

	// Mock expectations
	suite.debtListRepo.On("BelongsToUser", ctx, debtListID, userID).Return(true, nil)
	suite.debtListRepo.On("GetByID", ctx, debtListID).Return(debtList, nil)
	suite.debtItemRepo.On("GetCompletedPaymentsForDebtList", ctx, debtListID).Return(payments, nil)

	// Execute
	schedule, err := suite.debtService.GetPaymentSchedule(ctx, debtListID, userID)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, schedule)

	// Only first payment should be marked as paid
	assert.Equal(t, "paid", schedule[0].Status)
	assert.True(t, schedule[0].PaidAmount.Equal(decimal.RequireFromString("250.00")))

	// Rest should be pending
	for i := 1; i < len(schedule); i++ {
		assert.Equal(t, "pending", schedule[i].Status)
		assert.True(t, schedule[i].PaidAmount.IsZero())
	}

	suite.debtListRepo.AssertExpectations(t)
	suite.debtItemRepo.AssertExpectations(t)
}

