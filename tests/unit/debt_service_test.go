package unit

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"pay-your-dues/internal/domain/entities"
	"pay-your-dues/internal/mocks"
	"pay-your-dues/internal/services"
)

func TestDebtService_CreateDebtList(t *testing.T) {
	userID := uuid.New()
	contactID := uuid.New()
	futureDate := time.Now().AddDate(0, 0, 30)

	tests := []struct {
		name           string
		userID         uuid.UUID
		request        *entities.CreateDebtListRequest
		setupMocks     func(*mocks.MockDebtListRepository, *mocks.MockDebtItemRepository, *mocks.MockContactRepository, *mocks.MockPaymentScheduleService)
		expectedError  error
		expectSuccess  bool
		validateResult func(*testing.T, *entities.DebtList)
	}{
		{
			name:   "create debt list with one-time payment",
			userID: userID,
			request: &entities.CreateDebtListRequest{
				ContactID:   contactID,
				DebtType:    "to_pay",
				TotalAmount: "1000.00",
				Currency:    "USD",
				DueDate:     &futureDate,
				InstallmentPlan: "onetime",
				Description: stringPtr("Loan from friend"),
			},
			setupMocks: func(debtListRepo *mocks.MockDebtListRepository, debtItemRepo *mocks.MockDebtItemRepository, contactRepo *mocks.MockContactRepository, paymentService *mocks.MockPaymentScheduleService) {
				userContact := &entities.UserContact{
					ID:        uuid.New(),
					UserID:    userID,
					ContactID: contactID,
				}
				contactRepo.On("GetUserContactRelation", mock.Anything, userID, contactID).Return(userContact, nil)
				paymentService.On("CalculateInstallmentAmount", decimal.RequireFromString("1000.00"), "onetime", mock.AnythingOfType("time.Time"), futureDate).Return(decimal.RequireFromString("1000.00"))
				debtListRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.DebtList")).Return(nil)
			},
			expectedError: nil,
			expectSuccess: true,
			validateResult: func(t *testing.T, debtList *entities.DebtList) {
				assert.Equal(t, userID, debtList.UserID)
				assert.Equal(t, contactID, debtList.ContactID)
				assert.Equal(t, "to_pay", debtList.DebtType)
				assert.Equal(t, "1000", debtList.TotalAmount.String())
				assert.Equal(t, "1000", debtList.InstallmentAmount.String())
				assert.Equal(t, "active", debtList.Status)
				assert.Equal(t, "USD", debtList.Currency)
			},
		},
		{
			name:   "create debt list with monthly payments",
			userID: userID,
			request: &entities.CreateDebtListRequest{
				ContactID:        contactID,
				DebtType:         "to_receive",
				TotalAmount:      "2400.00",
				Currency:         "USD",
				InstallmentPlan:  "monthly",
				NumberOfPayments: intPtr(12),
				Description:      stringPtr("Personal loan"),
			},
			setupMocks: func(debtListRepo *mocks.MockDebtListRepository, debtItemRepo *mocks.MockDebtItemRepository, contactRepo *mocks.MockContactRepository, paymentService *mocks.MockPaymentScheduleService) {
				userContact := &entities.UserContact{
					ID:        uuid.New(),
					UserID:    userID,
					ContactID: contactID,
				}
				contactRepo.On("GetUserContactRelation", mock.Anything, userID, contactID).Return(userContact, nil)
				calculatedDueDate := time.Now().AddDate(0, 12, 0)
				paymentService.On("CalculateDueDateFromNumberOfPayments", mock.AnythingOfType("time.Time"), 12, "monthly").Return(calculatedDueDate)
				paymentService.On("CalculateInstallmentAmountFromNumberOfPayments", decimal.RequireFromString("2400.00"), 12).Return(decimal.RequireFromString("200.00"))
				paymentService.On("CalculateNextPaymentDate", mock.AnythingOfType("*entities.DebtList"), (*time.Time)(nil)).Return(time.Now().AddDate(0, 1, 0))
				debtListRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.DebtList")).Return(nil)
			},
			expectedError: nil,
			expectSuccess: true,
			validateResult: func(t *testing.T, debtList *entities.DebtList) {
				assert.Equal(t, "to_receive", debtList.DebtType)
				assert.Equal(t, "2400", debtList.TotalAmount.String())
				assert.Equal(t, "200", debtList.InstallmentAmount.String())
				assert.Equal(t, 12, *debtList.NumberOfPayments)
				assert.Equal(t, "monthly", debtList.InstallmentPlan)
			},
		},
		{
			name:   "invalid debt type",
			userID: userID,
			request: &entities.CreateDebtListRequest{
				ContactID:   contactID,
				DebtType:    "invalid_type",
				TotalAmount: "500.00",
				DueDate:     &futureDate,
			},
			setupMocks:    func(debtListRepo *mocks.MockDebtListRepository, debtItemRepo *mocks.MockDebtItemRepository, contactRepo *mocks.MockContactRepository, paymentService *mocks.MockPaymentScheduleService) {},
			expectedError: entities.ErrInvalidDebtType,
			expectSuccess: false,
		},
		{
			name:   "invalid amount",
			userID: userID,
			request: &entities.CreateDebtListRequest{
				ContactID:   contactID,
				DebtType:    "to_pay",
				TotalAmount: "-100.00",
				DueDate:     &futureDate,
			},
			setupMocks: func(debtListRepo *mocks.MockDebtListRepository, debtItemRepo *mocks.MockDebtItemRepository, contactRepo *mocks.MockContactRepository, paymentService *mocks.MockPaymentScheduleService) {
				userContact := &entities.UserContact{
					ID:        uuid.New(),
					UserID:    userID,
					ContactID: contactID,
				}
				contactRepo.On("GetUserContactRelation", mock.Anything, userID, contactID).Return(userContact, nil)
			},
			expectedError: entities.ErrInvalidAmount,
			expectSuccess: false,
		},
		{
			name:   "due date in the past",
			userID: userID,
			request: &entities.CreateDebtListRequest{
				ContactID:   contactID,
				DebtType:    "to_pay",
				TotalAmount: "500.00",
				DueDate:     timePtr(time.Now().AddDate(0, 0, -1)),
			},
			setupMocks: func(debtListRepo *mocks.MockDebtListRepository, debtItemRepo *mocks.MockDebtItemRepository, contactRepo *mocks.MockContactRepository, paymentService *mocks.MockPaymentScheduleService) {
				userContact := &entities.UserContact{
					ID:        uuid.New(),
					UserID:    userID,
					ContactID: contactID,
				}
				contactRepo.On("GetUserContactRelation", mock.Anything, userID, contactID).Return(userContact, nil)
				paymentService.On("CalculateInstallmentAmount", decimal.RequireFromString("500.00"), "onetime", mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).Return(decimal.RequireFromString("500.00"))
			},
			expectedError: entities.ErrInvalidDueDate,
			expectSuccess: false,
		},
		{
			name:   "contact not found",
			userID: userID,
			request: &entities.CreateDebtListRequest{
				ContactID:   uuid.New(),
				DebtType:    "to_pay",
				TotalAmount: "500.00",
				DueDate:     &futureDate,
			},
			setupMocks: func(debtListRepo *mocks.MockDebtListRepository, debtItemRepo *mocks.MockDebtItemRepository, contactRepo *mocks.MockContactRepository, paymentService *mocks.MockPaymentScheduleService) {
				contactRepo.On("GetUserContactRelation", mock.Anything, userID, mock.AnythingOfType("uuid.UUID")).Return(nil, entities.ErrContactNotFound)
			},
			expectedError: entities.ErrContactNotFound,
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			debtListRepo := &mocks.MockDebtListRepository{}
			debtItemRepo := &mocks.MockDebtItemRepository{}
			contactRepo := &mocks.MockContactRepository{}
			paymentService := &mocks.MockPaymentScheduleService{}
			tt.setupMocks(debtListRepo, debtItemRepo, contactRepo, paymentService)

			// Create service
			fileStorageService := &mocks.MockFileStorageService{}
			debtService := services.NewDebtService(debtListRepo, debtItemRepo, contactRepo, paymentService, fileStorageService, nil)

			// Execute
			ctx := context.Background()
			result, err := debtService.CreateDebtList(ctx, tt.userID, tt.request)

			// Assert
			if tt.expectSuccess {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			} else {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Nil(t, result)
			}

			// Verify mock expectations
			debtListRepo.AssertExpectations(t)
			debtItemRepo.AssertExpectations(t)
			contactRepo.AssertExpectations(t)
			paymentService.AssertExpectations(t)
		})
	}
}

func TestDebtService_CreateDebtItem(t *testing.T) {
	userID := uuid.New()
	debtListID := uuid.New()
	paymentDate := time.Now()

	tests := []struct {
		name           string
		userID         uuid.UUID
		request        *entities.CreateDebtItemRequest
		setupMocks     func(*mocks.MockDebtListRepository, *mocks.MockDebtItemRepository, *mocks.MockContactRepository, *mocks.MockPaymentScheduleService)
		expectedError  error
		expectSuccess  bool
		validateResult func(*testing.T, *entities.DebtItem)
	}{
		{
			name:   "create payment successfully",
			userID: userID,
			request: &entities.CreateDebtItemRequest{
				DebtListID:    debtListID,
				Amount:        "200.00",
				Currency:      "USD",
				PaymentDate:   paymentDate,
				PaymentMethod: "bank_transfer",
				Description:   stringPtr("First installment"),
			},
			setupMocks: func(debtListRepo *mocks.MockDebtListRepository, debtItemRepo *mocks.MockDebtItemRepository, contactRepo *mocks.MockContactRepository, paymentService *mocks.MockPaymentScheduleService) {
				debtListRepo.On("BelongsToUser", mock.Anything, debtListID, userID).Return(true, nil)
				debtList := &entities.DebtList{
					ID:              debtListID,
					UserID:          userID,
					Currency:        "USD",
					TotalAmount:     decimal.RequireFromString("1000.00"),
					NextPaymentDate: time.Now().AddDate(0, 1, 0), // Future date
					CreatedAt:       time.Now(),
				}
				debtListRepo.On("GetByID", mock.Anything, debtListID).Return(debtList, nil)
				debtItemRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.DebtItem")).Return(nil)
				
				// Mock the update calls
				debtItemRepo.On("GetTotalPaidForDebtList", mock.Anything, debtListID).Return(decimal.RequireFromString("200.00"), nil)
				debtItemRepo.On("GetLastPaymentDate", mock.Anything, debtListID).Return(&paymentDate, nil)
				paymentService.On("CalculateNextPaymentDate", mock.AnythingOfType("*entities.DebtList"), &paymentDate).Return(time.Now().AddDate(0, 1, 0))
				debtListRepo.On("UpdatePaymentTotals", mock.Anything, debtListID, decimal.RequireFromString("200.00"), decimal.RequireFromString("800.00")).Return(nil)
				debtListRepo.On("UpdateStatus", mock.Anything, debtListID, "active").Return(nil)
				debtListRepo.On("UpdateNextPaymentDate", mock.Anything, debtListID, mock.AnythingOfType("time.Time")).Return(nil)
			},
			expectedError: nil,
			expectSuccess: true,
			validateResult: func(t *testing.T, debtItem *entities.DebtItem) {
				assert.Equal(t, debtListID, debtItem.DebtListID)
				assert.Equal(t, "200", debtItem.Amount.String())
				assert.Equal(t, "USD", debtItem.Currency)
				assert.Equal(t, "bank_transfer", debtItem.PaymentMethod)
				assert.Equal(t, "completed", debtItem.Status)
			},
		},
		{
			name:   "invalid amount",
			userID: userID,
			request: &entities.CreateDebtItemRequest{
				DebtListID:    debtListID,
				Amount:        "0.00",
				PaymentDate:   paymentDate,
				PaymentMethod: "cash",
			},
			setupMocks: func(debtListRepo *mocks.MockDebtListRepository, debtItemRepo *mocks.MockDebtItemRepository, contactRepo *mocks.MockContactRepository, paymentService *mocks.MockPaymentScheduleService) {
				debtListRepo.On("BelongsToUser", mock.Anything, debtListID, userID).Return(true, nil)
				debtList := &entities.DebtList{
					ID:       debtListID,
					UserID:   userID,
					Currency: "USD",
				}
				debtListRepo.On("GetByID", mock.Anything, debtListID).Return(debtList, nil)
			},
			expectedError: entities.ErrInvalidAmount,
			expectSuccess: false,
		},
		{
			name:   "invalid payment method",
			userID: userID,
			request: &entities.CreateDebtItemRequest{
				DebtListID:    debtListID,
				Amount:        "100.00",
				PaymentDate:   paymentDate,
				PaymentMethod: "invalid_method",
			},
			setupMocks: func(debtListRepo *mocks.MockDebtListRepository, debtItemRepo *mocks.MockDebtItemRepository, contactRepo *mocks.MockContactRepository, paymentService *mocks.MockPaymentScheduleService) {
				// No mocks needed since validation should fail early
			},
			expectedError: entities.ErrInvalidPaymentMethod,
			expectSuccess: false,
		},
		{
			name:   "debt list not found",
			userID: userID,
			request: &entities.CreateDebtItemRequest{
				DebtListID:    uuid.New(),
				Amount:        "100.00",
				PaymentDate:   paymentDate,
				PaymentMethod: "cash",
			},
			setupMocks: func(debtListRepo *mocks.MockDebtListRepository, debtItemRepo *mocks.MockDebtItemRepository, contactRepo *mocks.MockContactRepository, paymentService *mocks.MockPaymentScheduleService) {
				debtListRepo.On("BelongsToUser", mock.Anything, mock.AnythingOfType("uuid.UUID"), userID).Return(false, nil)
				// Mock GetDebtListsWhereUserIsContact to return empty list (user is not a contact)
				debtListRepo.On("GetDebtListsWhereUserIsContact", mock.Anything, userID).Return([]entities.DebtListResponse{}, nil)
			},
			expectedError: entities.ErrDebtListNotFound,
			expectSuccess: false,
		},
		{
			name:   "create payment as contact successfully",
			userID: userID,
			request: &entities.CreateDebtItemRequest{
				DebtListID:    debtListID,
				Amount:        "150.00",
				Currency:      "EUR",
				PaymentDate:   paymentDate,
				PaymentMethod: "cash",
				Description:   stringPtr("Partial payment"),
			},
			setupMocks: func(debtListRepo *mocks.MockDebtListRepository, debtItemRepo *mocks.MockDebtItemRepository, contactRepo *mocks.MockContactRepository, paymentService *mocks.MockPaymentScheduleService) {
				// User doesn't own the debt list
				debtListRepo.On("BelongsToUser", mock.Anything, debtListID, userID).Return(false, nil)
				
				// User is a contact in the debt list
				contactDebtList := entities.DebtListResponse{
					ID:       debtListID,
					UserID:   uuid.New(), // Different user owns it
					DebtType: "to_receive", // From the owner's perspective
				}
				debtListRepo.On("GetDebtListsWhereUserIsContact", mock.Anything, userID).Return([]entities.DebtListResponse{contactDebtList}, nil)
				
				// Get the debt list for currency and other details
				debtList := &entities.DebtList{
					ID:              debtListID,
					UserID:          uuid.New(), // Different user owns it
					Currency:        "EUR",
					DebtType:        "to_receive", // From the owner's perspective
					TotalAmount:     decimal.RequireFromString("500.00"),
					NextPaymentDate: time.Now().AddDate(0, 1, 0),
					CreatedAt:       time.Now(),
				}
				debtListRepo.On("GetByID", mock.Anything, debtListID).Return(debtList, nil)
				
				// Mock debt item creation
				debtItemRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.DebtItem")).Return(nil)
				
				// Mock the update calls
				debtItemRepo.On("GetTotalPaidForDebtList", mock.Anything, debtListID).Return(decimal.RequireFromString("150.00"), nil)
				debtItemRepo.On("GetLastPaymentDate", mock.Anything, debtListID).Return(&paymentDate, nil)
				paymentService.On("CalculateNextPaymentDate", mock.AnythingOfType("*entities.DebtList"), &paymentDate).Return(time.Now().AddDate(0, 1, 0))
				debtListRepo.On("UpdatePaymentTotals", mock.Anything, debtListID, decimal.RequireFromString("150.00"), decimal.RequireFromString("350.00")).Return(nil)
				debtListRepo.On("UpdateStatus", mock.Anything, debtListID, "active").Return(nil)
				debtListRepo.On("UpdateNextPaymentDate", mock.Anything, debtListID, mock.AnythingOfType("time.Time")).Return(nil)
			},
			expectedError: nil,
			expectSuccess: true,
			validateResult: func(t *testing.T, debtItem *entities.DebtItem) {
				assert.Equal(t, debtListID, debtItem.DebtListID)
				assert.Equal(t, "150", debtItem.Amount.String())
				assert.Equal(t, "EUR", debtItem.Currency)
				assert.Equal(t, "cash", debtItem.PaymentMethod)
				// Since the user is a contact in an "to_receive" debt list, 
				// from their perspective it's "to_pay", so status should be "pending"
				assert.Equal(t, "pending", debtItem.Status)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			debtListRepo := &mocks.MockDebtListRepository{}
			debtItemRepo := &mocks.MockDebtItemRepository{}
			contactRepo := &mocks.MockContactRepository{}
			paymentService := &mocks.MockPaymentScheduleService{}
			tt.setupMocks(debtListRepo, debtItemRepo, contactRepo, paymentService)

			// Create service
			fileStorageService := &mocks.MockFileStorageService{}
			debtService := services.NewDebtService(debtListRepo, debtItemRepo, contactRepo, paymentService, fileStorageService, nil)

			// Execute
			ctx := context.Background()
			result, err := debtService.CreateDebtItem(ctx, tt.userID, tt.request)

			// Assert
			if tt.expectSuccess {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			} else {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Nil(t, result)
			}

			// Verify mock expectations
			debtListRepo.AssertExpectations(t)
			debtItemRepo.AssertExpectations(t)
			contactRepo.AssertExpectations(t)
			paymentService.AssertExpectations(t)
		})
	}
}

func TestDebtService_GetOverdueItems(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name          string
		userID        uuid.UUID
		setupMocks    func(*mocks.MockDebtListRepository, *mocks.MockDebtItemRepository, *mocks.MockContactRepository, *mocks.MockPaymentScheduleService)
		expectedError error
		expectSuccess bool
		expectedCount int
	}{
		{
			name:   "get overdue items successfully",
			userID: userID,
			setupMocks: func(debtListRepo *mocks.MockDebtListRepository, debtItemRepo *mocks.MockDebtItemRepository, contactRepo *mocks.MockContactRepository, paymentService *mocks.MockPaymentScheduleService) {
				overdueDebts := []entities.DebtList{
					{
						ID:              uuid.New(),
						UserID:          userID,
						DebtType:        "to_pay",
						TotalAmount:     decimal.RequireFromString("500.00"),
						NextPaymentDate: time.Now().AddDate(0, 0, -5),
						Status:          "overdue",
					},
					{
						ID:              uuid.New(),
						UserID:          userID,
						DebtType:        "to_receive",
						TotalAmount:     decimal.RequireFromString("300.00"),
						NextPaymentDate: time.Now().AddDate(0, 0, -10),
						Status:          "overdue",
					},
				}
				debtListRepo.On("GetOverdueForUser", mock.Anything, userID).Return(overdueDebts, nil)
			},
			expectedError: nil,
			expectSuccess: true,
			expectedCount: 2,
		},
		{
			name:   "no overdue items",
			userID: userID,
			setupMocks: func(debtListRepo *mocks.MockDebtListRepository, debtItemRepo *mocks.MockDebtItemRepository, contactRepo *mocks.MockContactRepository, paymentService *mocks.MockPaymentScheduleService) {
				debtListRepo.On("GetOverdueForUser", mock.Anything, userID).Return([]entities.DebtList{}, nil)
			},
			expectedError: nil,
			expectSuccess: true,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			debtListRepo := &mocks.MockDebtListRepository{}
			debtItemRepo := &mocks.MockDebtItemRepository{}
			contactRepo := &mocks.MockContactRepository{}
			paymentService := &mocks.MockPaymentScheduleService{}
			tt.setupMocks(debtListRepo, debtItemRepo, contactRepo, paymentService)

			// Create service
			fileStorageService := &mocks.MockFileStorageService{}
			debtService := services.NewDebtService(debtListRepo, debtItemRepo, contactRepo, paymentService, fileStorageService, nil)

			// Execute
			ctx := context.Background()
			result, err := debtService.GetOverdueItems(ctx, tt.userID)

			// Assert
			if tt.expectSuccess {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Len(t, result, tt.expectedCount)
			} else {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Nil(t, result)
			}

			// Verify mock expectations
			debtListRepo.AssertExpectations(t)
			debtItemRepo.AssertExpectations(t)
			contactRepo.AssertExpectations(t)
			paymentService.AssertExpectations(t)
		})
	}
}

// Helper functions
func intPtr(i int) *int {
	return &i
}

func timePtr(t time.Time) *time.Time {
	return &t
}
