package integration

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"pay-your-dues/internal/domain/entities"
	"pay-your-dues/internal/mocks"
	"pay-your-dues/internal/services"
)

// TestPaymentVerificationWorkflow tests the complete payment verification workflow
func TestPaymentVerificationWorkflow(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("Complete Verification Workflow - Debtor Creates Payment, Creditor Verifies", func(t *testing.T) {
		// Setup
		ctx := context.Background()
		
		// Create mock repositories
		mockDebtListRepo := new(mocks.MockDebtListRepository)
		mockDebtItemRepo := new(mocks.MockDebtItemRepository)
		mockContactRepo := new(mocks.MockContactRepository)
		mockPaymentScheduleService := new(mocks.MockPaymentScheduleService)
		mockFileStorageService := new(mocks.MockFileStorageService)

		// Create service
		debtService := services.NewDebtService(
			mockDebtListRepo,
			mockDebtItemRepo,
			mockContactRepo,
			mockPaymentScheduleService,
			mockFileStorageService,
			nil,
		)

		// Test data
		debtorID := uuid.New()
		creditorID := uuid.New()
		debtListID := uuid.New()
		debtItemID := uuid.New()
		receiptURL := "s3://bucket/receipts/test-receipt.jpg"
		
		futureDate := time.Now().AddDate(0, 1, 0) // 1 month from now
		
		debtList := &entities.DebtList{
			ID:                debtListID,
			UserID:            creditorID, // Creditor owns the debt list
			ContactID:         debtorID,   // Debtor is the contact
			DebtType:          "to_receive", // From creditor's perspective
			TotalAmount:       decimal.RequireFromString("1000.00"),
			InstallmentAmount: decimal.RequireFromString("250.00"),
			Currency:          "USD",
			Status:            "active",
			DueDate:           futureDate,
			NextPaymentDate:   futureDate, // Set next payment date to future to avoid overdue status
			InstallmentPlan:   "monthly",
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}

		// Step 1: Debtor creates a payment (should be pending)
		t.Run("Step 1: Debtor creates payment with receipt", func(t *testing.T) {
			createReq := &entities.CreateDebtItemRequest{
				DebtListID:      debtListID,
				Amount:          "250.00",
				Currency:        "USD",
				PaymentDate:     time.Now(),
				PaymentMethod:   "bank_transfer",
				Description:     stringPtr("First installment payment"),
				ReceiptPhotoURL: &receiptURL,
			}

			// Mock expectations for debtor creating payment
			mockDebtListRepo.On("BelongsToUser", ctx, debtListID, debtorID).Return(false, nil).Once()
			mockDebtListRepo.On("GetDebtListsWhereUserIsContact", ctx, debtorID).Return([]entities.DebtListResponse{
				{
					ID:       debtListID,
					UserID:   creditorID,
					DebtType: "to_receive",
				},
			}, nil).Once()
			mockDebtListRepo.On("GetByID", ctx, debtListID).Return(debtList, nil).Twice() // Called once for validation, once for update
			
			expectedDebtItem := &entities.DebtItem{
				ID:              debtItemID,
				DebtListID:      debtListID,
				Amount:          decimal.RequireFromString("250.00"),
				Currency:        "USD",
				PaymentDate:     time.Now(),
				PaymentMethod:   "bank_transfer",
				Description:     stringPtr("First installment payment"),
				Status:          entities.PaymentStatusPending, // Should be pending since debtor created it
				ReceiptPhotoURL: &receiptURL,
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			}
			
			mockDebtItemRepo.On("Create", ctx, mock.MatchedBy(func(item *entities.DebtItem) bool {
				return item.Status == entities.PaymentStatusPending &&
					item.Amount.Equal(decimal.RequireFromString("250.00"))
			})).Return(nil).Once()
			
			// Mock expectations for updating debt list totals (even for pending payments)
			mockDebtItemRepo.On("GetTotalPaidForDebtList", ctx, debtListID).Return(decimal.Zero, nil).Once()
			mockDebtItemRepo.On("GetLastPaymentDate", ctx, debtListID).Return((*time.Time)(nil), nil).Once()
			mockDebtListRepo.On("UpdatePaymentTotals", ctx, debtListID, decimal.Zero, decimal.RequireFromString("1000.00")).Return(nil).Once()
			mockDebtListRepo.On("UpdateStatus", ctx, debtListID, "active").Return(nil).Once()
			mockPaymentScheduleService.On("CalculateNextPaymentDate", debtList, (*time.Time)(nil)).Return(time.Now().AddDate(0, 1, 0)).Once()
			mockDebtListRepo.On("UpdateNextPaymentDate", ctx, debtListID, mock.AnythingOfType("time.Time")).Return(nil).Once()

			// Create the payment
			result, err := debtService.CreateDebtItem(ctx, debtorID, createReq)
			
			require.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, entities.PaymentStatusPending, result.Status)
			assert.Equal(t, receiptURL, *result.ReceiptPhotoURL)
			
			mockDebtListRepo.AssertExpectations(t)
			mockDebtItemRepo.AssertExpectations(t)
			mockPaymentScheduleService.AssertExpectations(t)
			
			// Store for next steps
			expectedDebtItem.ID = result.ID
		})

		// Step 2: Creditor views pending verifications
		t.Run("Step 2: Creditor views pending verifications", func(t *testing.T) {
			pendingPayment := entities.DebtItem{
				ID:              debtItemID,
				DebtListID:      debtListID,
				Amount:          decimal.RequireFromString("250.00"),
				Currency:        "USD",
				PaymentDate:     time.Now(),
				PaymentMethod:   "bank_transfer",
				Description:     stringPtr("First installment payment"),
				Status:          entities.PaymentStatusPending,
				ReceiptPhotoURL: &receiptURL,
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			}

			mockDebtItemRepo.On("GetPendingVerifications", ctx, creditorID).Return([]entities.DebtItem{pendingPayment}, nil).Once()

			// Get pending verifications
			pendingVerifications, err := debtService.GetPendingVerifications(ctx, creditorID)
			
			require.NoError(t, err)
			assert.Len(t, pendingVerifications, 1)
			assert.Equal(t, debtItemID, pendingVerifications[0].ID)
			assert.Equal(t, entities.PaymentStatusPending, pendingVerifications[0].Status)
			assert.NotNil(t, pendingVerifications[0].ReceiptPhotoURL)
			
			mockDebtItemRepo.AssertExpectations(t)
		})

		// Step 3: Creditor verifies the payment
		t.Run("Step 3: Creditor verifies payment", func(t *testing.T) {
			verifyReq := &entities.VerifyDebtItemRequest{
				Status:            entities.PaymentStatusCompleted,
				VerificationNotes: stringPtr("Payment verified, thank you!"),
			}

			pendingPayment := &entities.DebtItem{
				ID:              debtItemID,
				DebtListID:      debtListID,
				Amount:          decimal.RequireFromString("250.00"),
				Currency:        "USD",
				PaymentDate:     time.Now(),
				PaymentMethod:   "bank_transfer",
				Description:     stringPtr("First installment payment"),
				Status:          entities.PaymentStatusPending,
				ReceiptPhotoURL: &receiptURL,
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			}

			verifiedPayment := &entities.DebtItem{
				ID:                debtItemID,
				DebtListID:        debtListID,
				Amount:            decimal.RequireFromString("250.00"),
				Currency:          "USD",
				PaymentDate:       time.Now(),
				PaymentMethod:     "bank_transfer",
				Description:       stringPtr("First installment payment"),
				Status:            entities.PaymentStatusCompleted,
				ReceiptPhotoURL:   &receiptURL,
				VerifiedBy:        &creditorID,
				VerifiedAt:        timePtr(time.Now()),
				VerificationNotes: stringPtr("Payment verified, thank you!"),
				CreatedAt:         time.Now(),
				UpdatedAt:         time.Now(),
			}

			// Mock expectations for verification
			mockDebtItemRepo.On("CanUserVerifyDebtItem", ctx, debtItemID, creditorID).Return(true, nil).Twice()
			mockDebtItemRepo.On("GetByID", ctx, debtItemID).Return(pendingPayment, nil).Once()
			mockDebtItemRepo.On("UpdatePaymentStatus", ctx, debtItemID, entities.PaymentStatusCompleted, creditorID, verifyReq.VerificationNotes).Return(nil).Once()
			mockDebtItemRepo.On("GetByID", ctx, debtItemID).Return(verifiedPayment, nil).Once()
			
			// Mock expectations for updating debt list totals
			mockDebtListRepo.On("GetByID", ctx, debtListID).Return(debtList, nil).Once()
			mockDebtItemRepo.On("GetTotalPaidForDebtList", ctx, debtListID).Return(decimal.RequireFromString("250.00"), nil).Once()
			mockDebtItemRepo.On("GetLastPaymentDate", ctx, debtListID).Return(timePtr(time.Now()), nil).Once()
			mockDebtListRepo.On("UpdatePaymentTotals", ctx, debtListID, decimal.RequireFromString("250.00"), decimal.RequireFromString("750.00")).Return(nil).Once()
			mockDebtListRepo.On("UpdateStatus", ctx, debtListID, "active").Return(nil).Once()
			mockPaymentScheduleService.On("CalculateNextPaymentDate", debtList, mock.AnythingOfType("*time.Time")).Return(time.Now().AddDate(0, 1, 0)).Once()
			mockDebtListRepo.On("UpdateNextPaymentDate", ctx, debtListID, mock.AnythingOfType("time.Time")).Return(nil).Once()

			// Verify the payment
			result, err := debtService.VerifyDebtItem(ctx, debtItemID, creditorID, verifyReq)
			
			require.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, entities.PaymentStatusCompleted, result.Status)
			assert.NotNil(t, result.VerifiedBy)
			assert.Equal(t, creditorID, *result.VerifiedBy)
			assert.NotNil(t, result.VerifiedAt)
			assert.Equal(t, "Payment verified, thank you!", *result.VerificationNotes)
			
			mockDebtItemRepo.AssertExpectations(t)
			mockDebtListRepo.AssertExpectations(t)
			mockPaymentScheduleService.AssertExpectations(t)
		})
	})

	t.Run("Creditor Creates Payment - Auto Completed", func(t *testing.T) {
		// Setup
		ctx := context.Background()
		
		mockDebtListRepo := new(mocks.MockDebtListRepository)
		mockDebtItemRepo := new(mocks.MockDebtItemRepository)
		mockContactRepo := new(mocks.MockContactRepository)
		mockPaymentScheduleService := new(mocks.MockPaymentScheduleService)
		mockFileStorageService := new(mocks.MockFileStorageService)

		debtService := services.NewDebtService(
			mockDebtListRepo,
			mockDebtItemRepo,
			mockContactRepo,
			mockPaymentScheduleService,
			mockFileStorageService,
			nil,
		)

		creditorID := uuid.New()
		debtorID := uuid.New()
		debtListID := uuid.New()
		futureDate := time.Now().AddDate(0, 1, 0) // 1 month from now
		
		debtList := &entities.DebtList{
			ID:                debtListID,
			UserID:            creditorID,
			ContactID:         debtorID,
			DebtType:          "to_receive",
			TotalAmount:       decimal.RequireFromString("1000.00"),
			InstallmentAmount: decimal.RequireFromString("250.00"),
			Currency:          "USD",
			Status:            "active",
			DueDate:           futureDate,
			NextPaymentDate:   futureDate, // Set next payment date to future to avoid overdue status
			InstallmentPlan:   "monthly",
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}

		createReq := &entities.CreateDebtItemRequest{
			DebtListID:    debtListID,
			Amount:        "250.00",
			Currency:      "USD",
			PaymentDate:   time.Now(),
			PaymentMethod: "cash",
			Description:   stringPtr("Cash payment received"),
		}

		// Mock expectations - creditor owns the debt list
		mockDebtListRepo.On("BelongsToUser", ctx, debtListID, creditorID).Return(true, nil).Once()
		mockDebtListRepo.On("GetByID", ctx, debtListID).Return(debtList, nil).Twice()
		mockDebtItemRepo.On("Create", ctx, mock.MatchedBy(func(item *entities.DebtItem) bool {
			// When creditor creates payment, it should be completed immediately
			return item.Status == entities.PaymentStatusCompleted &&
				item.Amount.Equal(decimal.RequireFromString("250.00"))
		})).Return(nil).Once()
		
		// Mock expectations for updating debt list totals
		mockDebtItemRepo.On("GetTotalPaidForDebtList", ctx, debtListID).Return(decimal.RequireFromString("250.00"), nil).Once()
		mockDebtItemRepo.On("GetLastPaymentDate", ctx, debtListID).Return(timePtr(time.Now()), nil).Once()
		mockDebtListRepo.On("UpdatePaymentTotals", ctx, debtListID, decimal.RequireFromString("250.00"), decimal.RequireFromString("750.00")).Return(nil).Once()
		mockDebtListRepo.On("UpdateStatus", ctx, debtListID, "active").Return(nil).Once()
		mockPaymentScheduleService.On("CalculateNextPaymentDate", debtList, mock.AnythingOfType("*time.Time")).Return(time.Now().AddDate(0, 1, 0)).Once()
		mockDebtListRepo.On("UpdateNextPaymentDate", ctx, debtListID, mock.AnythingOfType("time.Time")).Return(nil).Once()

		// Create payment as creditor
		result, err := debtService.CreateDebtItem(ctx, creditorID, createReq)
		
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, entities.PaymentStatusCompleted, result.Status, "Payment should be auto-completed when creditor creates it")
		
		mockDebtListRepo.AssertExpectations(t)
		mockDebtItemRepo.AssertExpectations(t)
		mockPaymentScheduleService.AssertExpectations(t)
	})

	t.Run("Creditor Rejects Payment", func(t *testing.T) {
		// Setup
		ctx := context.Background()
		
		mockDebtListRepo := new(mocks.MockDebtListRepository)
		mockDebtItemRepo := new(mocks.MockDebtItemRepository)
		mockContactRepo := new(mocks.MockContactRepository)
		mockPaymentScheduleService := new(mocks.MockPaymentScheduleService)
		mockFileStorageService := new(mocks.MockFileStorageService)

		debtService := services.NewDebtService(
			mockDebtListRepo,
			mockDebtItemRepo,
			mockContactRepo,
			mockPaymentScheduleService,
			mockFileStorageService,
			nil,
		)

		creditorID := uuid.New()
		debtItemID := uuid.New()
		debtListID := uuid.New()
		receiptURL := "s3://bucket/receipts/test-receipt.jpg"
		rejectionNotes := "Receipt is not clear, please resubmit"

		pendingPayment := &entities.DebtItem{
			ID:              debtItemID,
			DebtListID:      debtListID,
			Amount:          decimal.RequireFromString("250.00"),
			Currency:        "USD",
			PaymentDate:     time.Now(),
			PaymentMethod:   "bank_transfer",
			Status:          entities.PaymentStatusPending,
			ReceiptPhotoURL: &receiptURL,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}

		rejectedPayment := &entities.DebtItem{
			ID:                debtItemID,
			DebtListID:        debtListID,
			Amount:            decimal.RequireFromString("250.00"),
			Currency:          "USD",
			PaymentDate:       time.Now(),
			PaymentMethod:     "bank_transfer",
			Status:            entities.PaymentStatusRejected,
			ReceiptPhotoURL:   &receiptURL,
			VerifiedBy:        &creditorID,
			VerifiedAt:        timePtr(time.Now()),
			VerificationNotes: &rejectionNotes,
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}

		// Mock expectations
		mockDebtItemRepo.On("CanUserVerifyDebtItem", ctx, debtItemID, creditorID).Return(true, nil).Twice()
		mockDebtItemRepo.On("GetByID", ctx, debtItemID).Return(pendingPayment, nil).Once()
		mockDebtItemRepo.On("UpdatePaymentStatus", ctx, debtItemID, entities.PaymentStatusRejected, creditorID, &rejectionNotes).Return(nil).Once()
		mockDebtItemRepo.On("GetByID", ctx, debtItemID).Return(rejectedPayment, nil).Once()

		// Reject the payment
		result, err := debtService.RejectDebtItem(ctx, debtItemID, creditorID, &rejectionNotes)
		
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, entities.PaymentStatusRejected, result.Status)
		assert.NotNil(t, result.VerifiedBy)
		assert.Equal(t, creditorID, *result.VerifiedBy)
		assert.NotNil(t, result.VerificationNotes)
		assert.Equal(t, rejectionNotes, *result.VerificationNotes)
		
		mockDebtItemRepo.AssertExpectations(t)
	})

	t.Run("Unauthorized User Cannot Verify Payment", func(t *testing.T) {
		// Setup
		ctx := context.Background()
		
		mockDebtListRepo := new(mocks.MockDebtListRepository)
		mockDebtItemRepo := new(mocks.MockDebtItemRepository)
		mockContactRepo := new(mocks.MockContactRepository)
		mockPaymentScheduleService := new(mocks.MockPaymentScheduleService)
		mockFileStorageService := new(mocks.MockFileStorageService)

		debtService := services.NewDebtService(
			mockDebtListRepo,
			mockDebtItemRepo,
			mockContactRepo,
			mockPaymentScheduleService,
			mockFileStorageService,
			nil,
		)

		unauthorizedUserID := uuid.New()
		debtItemID := uuid.New()

		verifyReq := &entities.VerifyDebtItemRequest{
			Status:            entities.PaymentStatusCompleted,
			VerificationNotes: stringPtr("Trying to verify"),
		}

		// Mock expectations - user cannot verify this payment
		mockDebtItemRepo.On("CanUserVerifyDebtItem", ctx, debtItemID, unauthorizedUserID).Return(false, nil).Once()

		// Attempt to verify
		result, err := debtService.VerifyDebtItem(ctx, debtItemID, unauthorizedUserID, verifyReq)
		
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, entities.ErrDebtItemNotFound, err)
		
		mockDebtItemRepo.AssertExpectations(t)
	})
}
