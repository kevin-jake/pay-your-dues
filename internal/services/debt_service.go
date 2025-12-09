package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"pay-your-dues/internal/domain/entities"
	"pay-your-dues/internal/domain/interfaces"
)

// debtService implements the DebtService interface
type debtService struct {
	debtListRepo           interfaces.DebtListRepository
	debtItemRepo           interfaces.DebtItemRepository
	contactRepo            interfaces.ContactRepository
	paymentScheduleService interfaces.PaymentScheduleService
	fileStorageService     interfaces.FileStorageService
	notificationService    interfaces.NotificationService
}

// NewDebtService creates a new debt service
func NewDebtService(
	debtListRepo interfaces.DebtListRepository,
	debtItemRepo interfaces.DebtItemRepository,
	contactRepo interfaces.ContactRepository,
	paymentScheduleService interfaces.PaymentScheduleService,
	fileStorageService interfaces.FileStorageService,
	notificationService interfaces.NotificationService,
) interfaces.DebtService {
	return &debtService{
		debtListRepo:           debtListRepo,
		debtItemRepo:           debtItemRepo,
		contactRepo:            contactRepo,
		paymentScheduleService: paymentScheduleService,
		fileStorageService:     fileStorageService,
		notificationService:    notificationService,
	}
}

// Debt List operations

func (s *debtService) CreateDebtList(ctx context.Context, userID uuid.UUID, req *entities.CreateDebtListRequest) (*entities.DebtList, error) {
	// Validate input
	if err := s.validateCreateDebtListRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Verify contact exists and belongs to user
	_, err := s.contactRepo.GetUserContactRelation(ctx, userID, req.ContactID)
	if err != nil {
		return nil, fmt.Errorf("contact verification failed: %w", err)
	}

	// Parse total amount
	totalAmount, err := decimal.NewFromString(req.TotalAmount)
	if err != nil {
		return nil, entities.ErrInvalidAmount
	}
	
	// Validate that amount is positive
	if totalAmount.LessThanOrEqual(decimal.Zero) {
		return nil, entities.ErrInvalidAmount
	}

	// Set default currency if not provided
	currency := req.Currency
	if currency == "" {
		currency = "Php"
	}

	// Validation: If number_of_payments is provided, installment_plan is required
	if req.NumberOfPayments != nil && *req.NumberOfPayments > 0 && req.InstallmentPlan == "" {
		return nil, fmt.Errorf("installment_plan is required when number_of_payments is provided")
	}

	// Validation: If due_date is provided but installment_plan is not, default to 1-time payment
	installmentPlan := req.InstallmentPlan
	if req.DueDate != nil && installmentPlan == "" {
		installmentPlan = "onetime" // Default to onetime for 1-time payment calculation
	}

	// Determine due date and installment amount based on input
	var dueDate time.Time
	var installmentAmount decimal.Decimal
	var numberOfPayments *int

	createdAt := time.Now()
	if req.NumberOfPayments != nil && *req.NumberOfPayments == 1 && req.DueDate != nil {
		dueDate = *req.DueDate
	} else if req.NumberOfPayments != nil && *req.NumberOfPayments > 0 {
		// Use number of payments to calculate due date and installment amount
		numberOfPayments = req.NumberOfPayments
		dueDate = s.paymentScheduleService.CalculateDueDateFromNumberOfPayments(createdAt, *req.NumberOfPayments, installmentPlan)
		installmentAmount = s.paymentScheduleService.CalculateInstallmentAmountFromNumberOfPayments(totalAmount, *req.NumberOfPayments)
	} else if req.DueDate != nil {
		// Use provided due date (existing behavior)
		dueDate = *req.DueDate
		installmentAmount = s.paymentScheduleService.CalculateInstallmentAmount(totalAmount, installmentPlan, createdAt, dueDate)
	} else {
		// Default to 1 payment if neither is provided
		defaultPayments := 1
		numberOfPayments = &defaultPayments
		dueDate = s.paymentScheduleService.CalculateDueDateFromNumberOfPayments(createdAt, defaultPayments, installmentPlan)
		installmentAmount = s.paymentScheduleService.CalculateInstallmentAmountFromNumberOfPayments(totalAmount, defaultPayments)
	}

	// Validate that due date is in the future
	now := time.Now()
	if dueDate.Before(now) || dueDate.Equal(now) {
		return nil, entities.ErrInvalidDueDate
	}

	// Calculate next payment date
	var nextPaymentDate time.Time
	if installmentPlan == "onetime" {
		nextPaymentDate = dueDate
	} else {
		nextPaymentDate = s.paymentScheduleService.CalculateNextPaymentDate(&entities.DebtList{
			InstallmentPlan: installmentPlan,
			CreatedAt:       createdAt,
			DueDate:         dueDate,
		}, nil)
	}

	debtList := &entities.DebtList{
		ID:                  uuid.New(),
		UserID:              userID,
		ContactID:           req.ContactID,
		DebtType:            req.DebtType,
		TotalAmount:         totalAmount,
		InstallmentAmount:   installmentAmount,
		TotalPaymentsMade:   decimal.Zero,
		TotalRemainingDebt:  totalAmount,
		Currency:            currency,
		Status:              "active",
		DueDate:             dueDate,
		NextPaymentDate:     nextPaymentDate,
		InstallmentPlan:     installmentPlan,
		NumberOfPayments:    numberOfPayments,
		Description:         req.Description,
		Notes:               req.Notes,
		CreatedAt:           createdAt,
		UpdatedAt:           createdAt,
	}

	// Validate debt list entity
	if err := debtList.IsValid(); err != nil {
		return nil, fmt.Errorf("invalid debt list entity: %w", err)
	}

	if err := s.debtListRepo.Create(ctx, debtList); err != nil {
		return nil, fmt.Errorf("failed to create debt list: %w", err)
	}

	// Schedule notifications for the new debt list (async, don't fail if notifications fail)
	if s.notificationService != nil {
		if err := s.notificationService.CreateNotificationsForDebtList(userID, debtList.ID); err != nil {
			// Log error but don't fail the debt list creation
			// TODO: Add proper logging here
			_ = err // Suppress unused variable warning for now
		}
	}

	return debtList, nil
}

func (s *debtService) GetDebtList(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*entities.DebtListResponse, error) {
	// First check if debt list belongs to user (user is the owner)
	belongs, err := s.debtListRepo.BelongsToUser(ctx, id, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify ownership: %w", err)
	}
	
	if belongs {
		// User owns the debt list, get it with relations
		debtListResponse, err := s.debtListRepo.GetByIDWithRelations(ctx, id, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to get debt list: %w", err)
		}
		return debtListResponse, nil
	}
	
	// If user doesn't own it, check if they are a contact in the debt list
	// This allows users to view debt lists where they owe money or are owed money
	contactDebtLists, err := s.debtListRepo.GetDebtListsWhereUserIsContact(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get debt lists where user is contact: %w", err)
	}
	
	// Find the specific debt list where user is a contact
	for _, contactDebtList := range contactDebtLists {
		if contactDebtList.ID == id {
			// Flip the debt type for the user's perspective
			// When a user views a debt list where they are the contact,
			// the debt type should be from their perspective
			if contactDebtList.DebtType == "to_receive" {
				contactDebtList.DebtType = "to_pay"
			} else if contactDebtList.DebtType == "to_pay" {
				contactDebtList.DebtType = "to_receive"
			}
			return &contactDebtList, nil
		}
	}
	
	// User neither owns the debt list nor is a contact in it
	return nil, entities.ErrDebtListNotFound
}

func (s *debtService) GetUserDebtLists(ctx context.Context, userID uuid.UUID) ([]entities.DebtListResponse, error) {
	// Get debt lists owned by the user
	ownedDebtLists, err := s.debtListRepo.GetUserDebtLists(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user debt lists: %w", err)
	}

	// Get debt lists where the user is referenced as a contact (other users owe them or they owe other users)
	contactDebtLists, err := s.debtListRepo.GetDebtListsWhereUserIsContact(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get debt lists where user is contact: %w", err)
	}

	// Flip the debt type for debt lists where the user is the contact
	for i := range contactDebtLists {
		if contactDebtLists[i].DebtType == "to_receive" {
			contactDebtLists[i].DebtType = "to_pay"
		} else if contactDebtLists[i].DebtType == "to_pay" {
			contactDebtLists[i].DebtType = "to_receive"
		}
	}

	// Combine both lists
	allDebtLists := append(ownedDebtLists, contactDebtLists...)

	return allDebtLists, nil
}

func (s *debtService) UpdateDebtList(ctx context.Context, id uuid.UUID, userID uuid.UUID, req *entities.UpdateDebtListRequest) (*entities.DebtList, error) {
	// Validate input
	if err := s.validateUpdateDebtListRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check if debt list belongs to user
	belongs, err := s.debtListRepo.BelongsToUser(ctx, id, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify ownership: %w", err)
	}
	if !belongs {
		return nil, entities.ErrDebtListNotFound
	}

	// Get existing debt list
	debtList, err := s.debtListRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get debt list: %w", err)
	}

	// Step 1: Update simple fields first
	if req.Currency != nil {
		debtList.Currency = *req.Currency
	}
	if req.Status != nil {
		debtList.Status = *req.Status
	}
	if req.Description != nil {
		debtList.Description = req.Description
	}
	if req.Notes != nil {
		debtList.Notes = req.Notes
	}

	// Step 2: Update InstallmentPlan (affects all calculations)
	if req.InstallmentPlan != nil {
		debtList.InstallmentPlan = *req.InstallmentPlan
		// For onetime, always set NumberOfPayments to 1
		if *req.InstallmentPlan == "onetime" {
			onePayment := 1
			debtList.NumberOfPayments = &onePayment
		}
	}

	// Step 3: Update TotalAmount
	if req.TotalAmount != nil {
		totalAmount, err := decimal.NewFromString(*req.TotalAmount)
		if err != nil {
			return nil, entities.ErrInvalidAmount
		}
		debtList.TotalAmount = totalAmount
	}

	// Step 4: Update NumberOfPayments (but respect onetime constraint)
	if req.NumberOfPayments != nil {
		if debtList.InstallmentPlan == "onetime" {
			// For onetime payment, always force to 1 regardless of what user provides
			onePayment := 1
			debtList.NumberOfPayments = &onePayment
		} else {
			debtList.NumberOfPayments = req.NumberOfPayments
		}
	}

	// Step 5: Update DueDate
	if req.DueDate != nil {
		// Only validate future date for non-onetime payments
		if debtList.InstallmentPlan != "onetime" {
			now := time.Now()
			if req.DueDate.Before(now) || req.DueDate.Equal(now) {
				return nil, entities.ErrInvalidDueDate
			}
		}
		debtList.DueDate = *req.DueDate
	}

	// Step 6: Recalculate dependent fields based on final state
	if debtList.InstallmentPlan == "onetime" {
		// Onetime payment: installment amount = total amount
		debtList.InstallmentAmount = debtList.TotalAmount
		// Ensure NumberOfPayments is 1
		if debtList.NumberOfPayments == nil {
			onePayment := 1
			debtList.NumberOfPayments = &onePayment
		}
	} else {
		// Non-onetime payment: calculate based on NumberOfPayments or DueDate
		if debtList.NumberOfPayments != nil && *debtList.NumberOfPayments > 0 {
			// Use NumberOfPayments to calculate installment amount and due date
			debtList.InstallmentAmount = s.paymentScheduleService.CalculateInstallmentAmountFromNumberOfPayments(debtList.TotalAmount, *debtList.NumberOfPayments)
			// Only recalculate due date if it wasn't explicitly set in this request
			if req.DueDate == nil {
				debtList.DueDate = s.paymentScheduleService.CalculateDueDateFromNumberOfPayments(debtList.CreatedAt, *debtList.NumberOfPayments, debtList.InstallmentPlan)
			}
		} else {
			// Use DueDate to calculate installment amount and number of payments
			debtList.InstallmentAmount = s.paymentScheduleService.CalculateInstallmentAmount(debtList.TotalAmount, debtList.InstallmentPlan, debtList.CreatedAt, debtList.DueDate)
			// Calculate number of payments if not explicitly set
			if debtList.NumberOfPayments == nil && debtList.InstallmentAmount.GreaterThan(decimal.Zero) {
				paymentsNeeded := debtList.TotalAmount.Div(debtList.InstallmentAmount).Ceil().IntPart()
				paymentsNeededInt := int(paymentsNeeded)
				debtList.NumberOfPayments = &paymentsNeededInt
			}
		}
	}

	debtList.UpdatedAt = time.Now()

	// Validate updated debt list entity
	if err := debtList.IsValid(); err != nil {
		return nil, fmt.Errorf("invalid updated debt list entity: %w", err)
	}

	// Save to database
	if err := s.debtListRepo.Update(ctx, debtList); err != nil {
		return nil, fmt.Errorf("failed to update debt list: %w", err)
	}

	// Recalculate payment totals and status based on actual debt items
	if err := s.updateDebtListStatusAndPaymentTotals(ctx, debtList.ID); err != nil {
		return nil, fmt.Errorf("failed to update payment totals: %w", err)
	}

	// Fetch the updated debt list to return the latest state
	updatedDebtList, err := s.debtListRepo.GetByID(ctx, debtList.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated debt list: %w", err)
	}

	return updatedDebtList, nil
}

func (s *debtService) DeleteDebtList(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	// Check if debt list belongs to user
	belongs, err := s.debtListRepo.BelongsToUser(ctx, id, userID)
	if err != nil {
		return fmt.Errorf("failed to verify ownership: %w", err)
	}
	if !belongs {
		return entities.ErrDebtListNotFound
	}

	if err := s.debtListRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete debt list: %w", err)
	}

	return nil
}

// Debt Item (Payment) operations

func (s *debtService) CreateDebtItem(ctx context.Context, userID uuid.UUID, req *entities.CreateDebtItemRequest) (*entities.DebtItem, error) {
	// Validate input
	if err := s.validateCreateDebtItemRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// First check if debt list belongs to user (user is the owner)
	belongs, err := s.debtListRepo.BelongsToUser(ctx, req.DebtListID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify debt list ownership: %w", err)
	}
	
	if belongs {
		// User owns the debt list, proceed with creation
	} else {
		// If user doesn't own it, check if they are a contact in the debt list
		// This allows users to create debt items for debt lists where they owe money or are owed money
		contactDebtLists, err := s.debtListRepo.GetDebtListsWhereUserIsContact(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to get debt lists where user is contact: %w", err)
		}
		
		// Check if the user is a contact in the specific debt list
		userIsContact := false
		for _, contactDebtList := range contactDebtLists {
			if contactDebtList.ID == req.DebtListID {
				userIsContact = true
				break
			}
		}
		
		if !userIsContact {
			// User neither owns the debt list nor is a contact in it
			return nil, entities.ErrDebtListNotFound
		}
	}

	// Get debt list for currency default
	debtList, err := s.debtListRepo.GetByID(ctx, req.DebtListID)
	if err != nil {
		return nil, fmt.Errorf("failed to get debt list: %w", err)
	}

	// Parse amount
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		return nil, entities.ErrInvalidAmount
	}
	
	// Validate that amount is positive
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, entities.ErrInvalidAmount
	}

	// Set default currency if not provided
	currency := req.Currency
	if currency == "" {
		currency = debtList.Currency
	}

	// Determine initial status based on the user's perspective
	// When a user creates a payment, we need to consider their perspective:
	// - If they owe money (to_pay from their perspective), payments are pending until verified
	// - If they are owed money (to_receive from their perspective), payments are completed
	initialStatus := "completed"
	
	// Check if the user is the owner or a contact to determine their perspective
	if belongs {
		// User owns the debt list, use the debt list's debt type
		if debtList.DebtType == "to_pay" {
			initialStatus = "pending"
		}
	} else {
		// User is a contact, determine their perspective by flipping the debt type
		// If the debt list is "to_receive" (someone owes them), then from the contact's perspective it's "to_pay"
		if debtList.DebtType == "to_receive" {
			initialStatus = "pending"
		}
	}

	debtItem := &entities.DebtItem{
		ID:                uuid.New(),
		DebtListID:        req.DebtListID,
		Amount:            amount,
		Currency:          currency,
		PaymentDate:       req.PaymentDate,
		PaymentMethod:     req.PaymentMethod,
		Description:       req.Description,
		Status:            initialStatus,
		ReceiptPhotoURL:   req.ReceiptPhotoURL,
		VerificationNotes: req.VerificationNotes,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	// Validate debt item entity
	if err := debtItem.IsValid(); err != nil {
		return nil, fmt.Errorf("invalid debt item entity: %w", err)
	}

	if err := s.debtItemRepo.Create(ctx, debtItem); err != nil {
		return nil, fmt.Errorf("failed to create debt item: %w", err)
	}

	// Update debt list status, next payment date, and payment totals
	if err := s.updateDebtListStatusAndPaymentTotals(ctx, debtList.ID); err != nil {
		return nil, fmt.Errorf("failed to update debt list totals: %w", err)
	}

	// Send payment confirmation notification (async, don't fail if notification fails)
	if s.notificationService != nil {
		if err := s.notificationService.SendPaymentConfirmationNotifications(debtItem.ID); err != nil {
			// Log error but don't fail the debt item creation
			_ = err // Suppress unused variable warning for now
		}
	}

	return debtItem, nil
}

func (s *debtService) GetDebtItem(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*entities.DebtItem, error) {
	// Check if debt item belongs to user's debt list
	belongs, err := s.debtItemRepo.BelongsToUserDebtList(ctx, id, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify ownership: %w", err)
	}
	if !belongs {
		return nil, entities.ErrDebtItemNotFound
	}

	debtItem, err := s.debtItemRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get debt item: %w", err)
	}

	return debtItem, nil
}

// GetDebtItemForVerification gets a debt item for verification purposes
// Only allows verification if the user is the one who should verify (debt_type = "to_receive")
func (s *debtService) GetDebtItemForVerification(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*entities.DebtItem, error) {
	// Check if user can verify this debt item
	canVerify, err := s.debtItemRepo.CanUserVerifyDebtItem(ctx, id, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify permission: %w", err)
	}
	if !canVerify {
		return nil, entities.ErrDebtItemNotFound
	}

	debtItem, err := s.debtItemRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get debt item: %w", err)
	}

	return debtItem, nil
}

func (s *debtService) GetDebtListItems(ctx context.Context, debtListID uuid.UUID, userID uuid.UUID) ([]entities.DebtItem, error) {
	// First check if debt list belongs to user (user is the owner)
	belongs, err := s.debtListRepo.BelongsToUser(ctx, debtListID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify ownership: %w", err)
	}
	
	if belongs {
		// User owns the debt list, get the debt items
		debtItems, err := s.debtItemRepo.GetByDebtListID(ctx, debtListID)
		if err != nil {
			return nil, fmt.Errorf("failed to get debt list items: %w", err)
		}
		return debtItems, nil
	}
	
	// If user doesn't own it, check if they are a contact in the debt list
	// This allows users to view debt items for debt lists where they owe money or are owed money
	contactDebtLists, err := s.debtListRepo.GetDebtListsWhereUserIsContact(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get debt lists where user is contact: %w", err)
	}
	
	// Check if the user is a contact in the specific debt list
	for _, contactDebtList := range contactDebtLists {
		if contactDebtList.ID == debtListID {
			// User is a contact in this debt list, get the debt items
			debtItems, err := s.debtItemRepo.GetByDebtListID(ctx, debtListID)
			if err != nil {
				return nil, fmt.Errorf("failed to get debt list items: %w", err)
			}
			return debtItems, nil
		}
	}
	
	// User neither owns the debt list nor is a contact in it
	return nil, entities.ErrDebtListNotFound
}

func (s *debtService) UpdateDebtItem(ctx context.Context, id uuid.UUID, userID uuid.UUID, req *entities.UpdateDebtItemRequest) (*entities.DebtItem, error) {
	// Validate input
	if err := s.validateUpdateDebtItemRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// First check if debt item belongs to user's debt list (user is the owner)
	belongs, err := s.debtItemRepo.BelongsToUserDebtList(ctx, id, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify ownership: %w", err)
	}
	
	if belongs {
		// User owns the debt list, proceed with update
	} else {
		// If user doesn't own it, check if they are a contact in the debt list
		// This allows users to update debt items for debt lists where they owe money or are owed money
		debtItem, err := s.debtItemRepo.GetByID(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("failed to get debt item: %w", err)
		}
		
		// Check if the user is a contact in this debt list
		contactDebtLists, err := s.debtListRepo.GetDebtListsWhereUserIsContact(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to get debt lists where user is contact: %w", err)
		}
		
		// Check if the user is a contact in the specific debt list
		userIsContact := false
		for _, contactDebtList := range contactDebtLists {
			if contactDebtList.ID == debtItem.DebtListID {
				userIsContact = true
				break
			}
		}
		
		if !userIsContact {
			// User neither owns the debt list nor is a contact in it
			return nil, entities.ErrDebtItemNotFound
		}
	}

	// Get existing debt item
	debtItem, err := s.debtItemRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get debt item: %w", err)
	}

	// Update fields if provided
	if req.Amount != nil {
		amount, err := decimal.NewFromString(*req.Amount)
		if err != nil {
			return nil, entities.ErrInvalidAmount
		}
		debtItem.Amount = amount
	}
	if req.Currency != nil {
		debtItem.Currency = *req.Currency
	}
	if req.PaymentDate != nil {
		debtItem.PaymentDate = *req.PaymentDate
	}
	if req.PaymentMethod != nil {
		debtItem.PaymentMethod = *req.PaymentMethod
	}
	if req.Description != nil {
		debtItem.Description = req.Description
	}
	if req.Status != nil {
		debtItem.Status = *req.Status
	}
	if req.ReceiptPhotoURL != nil {
		// If there's an old receipt photo, delete it from S3
		if debtItem.ReceiptPhotoURL != nil && *debtItem.ReceiptPhotoURL != "" {
			if err := s.fileStorageService.DeleteReceipt(ctx, *debtItem.ReceiptPhotoURL); err != nil {
				// Log the error but don't fail the update
				// In production, you might want to handle this differently
				fmt.Printf("Warning: failed to delete old receipt photo: %v\n", err)
			}
		}
		debtItem.ReceiptPhotoURL = req.ReceiptPhotoURL
	}
	if req.VerificationNotes != nil {
		debtItem.VerificationNotes = req.VerificationNotes
	}

	debtItem.UpdatedAt = time.Now()

	// Validate updated debt item entity
	if err := debtItem.IsValid(); err != nil {
		return nil, fmt.Errorf("invalid updated debt item entity: %w", err)
	}

	if err := s.debtItemRepo.Update(ctx, debtItem); err != nil {
		return nil, fmt.Errorf("failed to update debt item: %w", err)
	}

	// Update debt list status, next payment date, and payment totals
	if err := s.updateDebtListStatusAndPaymentTotals(ctx, debtItem.DebtListID); err != nil {
		return nil, fmt.Errorf("failed to update debt list totals: %w", err)
	}

	return debtItem, nil
}

func (s *debtService) DeleteDebtItem(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	// Check if debt item belongs to user's debt list
	belongs, err := s.debtItemRepo.BelongsToUserDebtList(ctx, id, userID)
	if err != nil {
		return fmt.Errorf("failed to verify ownership: %w", err)
	}
	if !belongs {
		return entities.ErrDebtItemNotFound
	}

	// Get debt item to know which debt list to update
	debtItem, err := s.debtItemRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get debt item: %w", err)
	}

	debtListID := debtItem.DebtListID

	// If there's a receipt photo, delete it from S3
	if debtItem.ReceiptPhotoURL != nil && *debtItem.ReceiptPhotoURL != "" {
		if err := s.fileStorageService.DeleteReceipt(ctx, *debtItem.ReceiptPhotoURL); err != nil {
			// Log the error but don't fail the deletion
			// In production, you might want to handle this differently
			fmt.Printf("Warning: failed to delete receipt photo: %v\n", err)
		}
	}

	if err := s.debtItemRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete debt item: %w", err)
	}

	// Update debt list status, next payment date, and payment totals
	if err := s.updateDebtListStatusAndPaymentTotals(ctx, debtListID); err != nil {
		return fmt.Errorf("failed to update debt list totals: %w", err)
	}

	return nil
}

// Debt analytics and reporting

func (s *debtService) GetOverdueItems(ctx context.Context, userID uuid.UUID) ([]entities.DebtList, error) {
	debtLists, err := s.debtListRepo.GetOverdueForUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get overdue items: %w", err)
	}

	return debtLists, nil
}

func (s *debtService) GetDueSoonItems(ctx context.Context, userID uuid.UUID, days int) ([]entities.DebtList, error) {
	dueDate := time.Now().AddDate(0, 0, days)
	debtLists, err := s.debtListRepo.GetDueSoonForUser(ctx, userID, dueDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get due soon items: %w", err)
	}

	return debtLists, nil
}

func (s *debtService) GetPaymentSchedule(ctx context.Context, debtListID uuid.UUID, userID uuid.UUID) ([]entities.PaymentScheduleItem, error) {
	// Check if debt list belongs to user
	belongs, err := s.debtListRepo.BelongsToUser(ctx, debtListID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify ownership: %w", err)
	}
	if !belongs {
		isContact, err := s.debtListRepo.IsContactOfDebtList(ctx, debtListID, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to verify contact association: %w", err)
		}
		if !isContact {
			return nil, entities.ErrDebtListNotFound
		}
	}

	// Get debt list
	debtList, err := s.debtListRepo.GetByID(ctx, debtListID)
	if err != nil {
		return nil, fmt.Errorf("failed to get debt list: %w", err)
	}

	// Get payments
	payments, err := s.debtItemRepo.GetCompletedPaymentsForDebtList(ctx, debtListID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payments: %w", err)
	}

	schedule := s.paymentScheduleService.CalculatePaymentSchedule(debtList, payments)
	return schedule, nil
}

func (s *debtService) GetUpcomingPayments(ctx context.Context, userID uuid.UUID, days int) ([]entities.UpcomingPayment, error) {
	debtLists, err := s.debtListRepo.GetUserDebtLists(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user debt lists: %w", err)
	}

	var upcomingPayments []entities.UpcomingPayment
	now := time.Now()
	cutoffDate := now.AddDate(0, 0, days)

	for _, debtList := range debtLists {
		// Skip settled or archived debts
		if debtList.Status == "settled" || debtList.Status == "archived" {
			continue
		}
		
		// Only consider active or overdue debts with remaining balance
		if (debtList.Status == "active" || debtList.Status == "overdue") && debtList.TotalRemainingDebt.GreaterThan(decimal.Zero) {
			var paymentDate time.Time
			var amountToPay decimal.Decimal
			
			// Check if the debt list's final due date is overdue
			// If so, prioritize showing that over individual installments
			if debtList.DueDate.Before(now) {
				// Debt list due date is overdue - show this instead of installments
				paymentDate = debtList.DueDate
				amountToPay = debtList.TotalRemainingDebt
			} else if debtList.InstallmentPlan == "onetime" {
				// For onetime payments, use the due date
				paymentDate = debtList.DueDate
				// Amount is the remaining debt (considering completed payments)
				amountToPay = debtList.TotalRemainingDebt
			} else {
				// For recurring payments with non-overdue final due date, get next installment
				// Get completed payments for this debt list
				payments, err := s.debtItemRepo.GetCompletedPaymentsForDebtList(ctx, debtList.ID)
				if err != nil {
					// If we can't get payments, skip this debt
					continue
				}
				
				// Convert DebtListResponse to DebtList for schedule calculation
				debtListEntity := &entities.DebtList{
					ID:                  debtList.ID,
					UserID:              debtList.UserID,
					ContactID:           debtList.ContactID,
					DebtType:            debtList.DebtType,
					TotalAmount:         debtList.TotalAmount,
					InstallmentAmount:   debtList.InstallmentAmount,
					TotalPaymentsMade:   debtList.TotalPaymentsMade,
					TotalRemainingDebt:  debtList.TotalRemainingDebt,
					Currency:            debtList.Currency,
					Status:              debtList.Status,
					DueDate:             debtList.DueDate,
					NextPaymentDate:     debtList.NextPaymentDate,
					InstallmentPlan:     debtList.InstallmentPlan,
					NumberOfPayments:    debtList.NumberOfPayments,
					Description:         debtList.Description,
					Notes:               debtList.Notes,
					CreatedAt:           debtList.CreatedAt,
					UpdatedAt:           debtList.UpdatedAt,
				}
				
				// Calculate payment schedule
				schedule := s.paymentScheduleService.CalculatePaymentSchedule(debtListEntity, payments)
				
				// Find the first pending payment in the schedule
				var nextScheduleItem *entities.PaymentScheduleItem
				for i := range schedule {
					if schedule[i].Status == "pending" || schedule[i].Status == "overdue" {
						nextScheduleItem = &schedule[i]
						break
					}
				}
				
				// If no pending payment found, skip this debt
				if nextScheduleItem == nil {
					continue
				}
				
				paymentDate = nextScheduleItem.DueDate
				amountToPay = nextScheduleItem.Amount
			}
			// Calculate days until due (negative if overdue, positive if upcoming), based on date only
			// Strip the time component from both dates to compare by date
			y, m, d := now.Date()
			nowDate := time.Date(y, m, d, 0, 0, 0, 0, now.Location())
			py, pm, pd := paymentDate.Date()
			paymentDateOnly := time.Date(py, pm, pd, 0, 0, 0, 0, paymentDate.Location())
			daysUntilDue := int(paymentDateOnly.Sub(nowDate).Hours() / 24)
			
			// Include payment if it's overdue OR within the upcoming range
			isOverdue := paymentDate.Before(now)
			isUpcoming := paymentDate.After(now) && (paymentDate.Before(cutoffDate) || paymentDate.Equal(cutoffDate))
			
			if isOverdue || isUpcoming {
				// Get contact name from UserContact (user-specific)
				userContact, err := s.contactRepo.GetUserContactRelation(ctx, userID, debtList.ContactID)
				contactName := "Unknown"
				if err == nil && userContact != nil {
					contactName = userContact.Name
				}
				
				upcomingPayment := entities.UpcomingPayment{
					DebtListID:      debtList.ID,
					ContactName:     contactName,
					DebtType:        debtList.DebtType,
					NextPaymentDate: paymentDate,
					DaysUntilDue:    daysUntilDue,
					Amount:          amountToPay,
					Currency:        debtList.Currency,
					Description:     debtList.Description,
				}
				upcomingPayments = append(upcomingPayments, upcomingPayment)
			}
		}
	}

	return upcomingPayments, nil
}

func (s *debtService) GetTotalPaymentsForDebtList(ctx context.Context, debtListID uuid.UUID, userID uuid.UUID) (*entities.PaymentSummary, error) {
	// Check if debt list belongs to user
	belongs, err := s.debtListRepo.BelongsToUser(ctx, debtListID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify ownership: %w", err)
	}
	if !belongs {
		return nil, entities.ErrDebtListNotFound
	}

	// Get debt list
	debtList, err := s.debtListRepo.GetByID(ctx, debtListID)
	if err != nil {
		return nil, fmt.Errorf("failed to get debt list: %w", err)
	}

	// Get all completed payments for this debt list
	payments, err := s.debtItemRepo.GetCompletedPaymentsForDebtList(ctx, debtListID)
	if err != nil {
		return nil, fmt.Errorf("failed to get completed payments: %w", err)
	}

	// Calculate total payments made
	totalPaid := decimal.Zero
	for _, payment := range payments {
		totalPaid = totalPaid.Add(payment.Amount)
	}

	// Calculate remaining debt
	remainingDebt := debtList.TotalAmount.Sub(totalPaid)
	if remainingDebt.LessThan(decimal.Zero) {
		remainingDebt = decimal.Zero
	}

	// Calculate percentage paid
	percentagePaid := decimal.Zero
	if debtList.TotalAmount.GreaterThan(decimal.Zero) {
		percentagePaid = totalPaid.Div(debtList.TotalAmount).Mul(decimal.NewFromInt(100))
	}

	return &entities.PaymentSummary{
		DebtListID:       debtListID,
		TotalAmount:      debtList.TotalAmount,
		TotalPaid:        totalPaid,
		RemainingDebt:    remainingDebt,
		PercentagePaid:   percentagePaid,
		NumberOfPayments: len(payments),
		Payments:         payments,
	}, nil
}

// Payment verification operations

func (s *debtService) VerifyDebtItem(ctx context.Context, id uuid.UUID, userID uuid.UUID, req *entities.VerifyDebtItemRequest) (*entities.DebtItem, error) {
	// Get the debt item to check if it exists and user can verify it
	debtItem, err := s.GetDebtItemForVerification(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	// Update the payment status and verification details
	if err := s.debtItemRepo.UpdatePaymentStatus(ctx, id, req.Status, userID, req.VerificationNotes); err != nil {
		return nil, fmt.Errorf("failed to update payment status: %w", err)
	}

	// Get the updated debt item
	updatedDebtItem, err := s.GetDebtItemForVerification(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	// Update debt list totals if payment is completed
	if req.Status == entities.PaymentStatusCompleted {
		if err := s.updateDebtListStatusAndPaymentTotals(ctx, debtItem.DebtListID); err != nil {
			return nil, fmt.Errorf("failed to update debt list totals: %w", err)
		}

		// Auto-disable notifications for paid installments
		if s.notificationService != nil {
			// Get the debt list to check payment schedule
			debtListForSchedule, err := s.debtListRepo.GetByID(ctx, debtItem.DebtListID)
			if err == nil {
				// Get all debt items for this debt list
				debtItems, err := s.debtItemRepo.GetByDebtListID(ctx, debtItem.DebtListID)
				if err == nil {
					// Calculate payment schedule to determine which installment was paid
					schedule := s.paymentScheduleService.CalculatePaymentSchedule(debtListForSchedule, debtItems)
					// Find paid installments and disable their notifications
					for _, item := range schedule {
						if item.Status == "paid" {
							if err := s.notificationService.DisableNotificationsForPaidInstallment(debtItem.DebtListID, item.PaymentNumber); err != nil {
								// Log error but don't fail the verification
								_ = err
							}
						}
					}
				}
			}
		}
	}

	// Send verification notification (async, don't fail if notification fails)
	if s.notificationService != nil {
		verified := req.Status == entities.PaymentStatusCompleted
		reason := ""
		if req.VerificationNotes != nil {
			reason = *req.VerificationNotes
		}
		if err := s.notificationService.SendPaymentVerificationNotification(updatedDebtItem.ID, verified, reason); err != nil {
			// Log error but don't fail the verification
			_ = err // Suppress unused variable warning for now
		}
	}

	return updatedDebtItem, nil
}

func (s *debtService) GetPendingVerifications(ctx context.Context, userID uuid.UUID) ([]entities.DebtItem, error) {
	return s.debtItemRepo.GetPendingVerifications(ctx, userID)
}

func (s *debtService) RejectDebtItem(ctx context.Context, id uuid.UUID, userID uuid.UUID, notes *string) (*entities.DebtItem, error) {
	// Check if the debt item exists and user can verify it
	_, err := s.GetDebtItemForVerification(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	// Update the payment status to rejected
	if err := s.debtItemRepo.UpdatePaymentStatus(ctx, id, entities.PaymentStatusRejected, userID, notes); err != nil {
		return nil, fmt.Errorf("failed to reject payment: %w", err)
	}

	// Get the updated debt item
	updatedDebtItem, err := s.GetDebtItemForVerification(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	// Send rejection notification (async, don't fail if notification fails)
	if s.notificationService != nil {
		reason := ""
		if notes != nil {
			reason = *notes
		}
		if err := s.notificationService.SendPaymentVerificationNotification(updatedDebtItem.ID, false, reason); err != nil {
			// Log error but don't fail the rejection
			_ = err // Suppress unused variable warning for now
		}
	}

	return updatedDebtItem, nil
}

// Helper methods

func (s *debtService) updateDebtListStatusAndPaymentTotals(ctx context.Context, debtListID uuid.UUID) error {
	// Get the debt list
	debtList, err := s.debtListRepo.GetByID(ctx, debtListID)
	if err != nil {
		return fmt.Errorf("failed to get debt list: %w", err)
	}

	// Calculate total payments made
	totalPaid, err := s.debtItemRepo.GetTotalPaidForDebtList(ctx, debtListID)
	if err != nil {
		return fmt.Errorf("failed to get total paid: %w", err)
	}

	// Calculate remaining amount
	remainingAmount := debtList.TotalAmount.Sub(totalPaid)
	if remainingAmount.LessThan(decimal.Zero) {
		remainingAmount = decimal.Zero
	}

	// Determine new status
	var newStatus string
	if remainingAmount.LessThanOrEqual(decimal.Zero) {
		newStatus = "settled"
	} else if time.Now().After(debtList.NextPaymentDate) {
		newStatus = "overdue"
	} else {
		newStatus = "active"
	}

	// Get the last payment date to calculate next payment
	lastPaymentDate, err := s.debtItemRepo.GetLastPaymentDate(ctx, debtListID)
	if err != nil {
		return fmt.Errorf("failed to get last payment date: %w", err)
	}

	// Calculate next payment date
	nextPaymentDate := s.paymentScheduleService.CalculateNextPaymentDate(debtList, lastPaymentDate)

	// Update debt list totals
	if err := s.debtListRepo.UpdatePaymentTotals(ctx, debtListID, totalPaid, remainingAmount); err != nil {
		return fmt.Errorf("failed to update payment totals: %w", err)
	}

	// Update status
	if err := s.debtListRepo.UpdateStatus(ctx, debtListID, newStatus); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	// Update next payment date
	if err := s.debtListRepo.UpdateNextPaymentDate(ctx, debtListID, nextPaymentDate); err != nil {
		return fmt.Errorf("failed to update next payment date: %w", err)
	}

	return nil
}

// Validation methods

func (s *debtService) validateCreateDebtListRequest(req *entities.CreateDebtListRequest) error {
	if req.ContactID == uuid.Nil {
		return entities.ErrInvalidInput
	}
	if req.DebtType != "to_receive" && req.DebtType != "to_pay" {
		return entities.ErrInvalidDebtType
	}
	if req.TotalAmount == "" {
		return entities.ErrInvalidAmount
	}
	// Validate that either due_date or number_of_payments is provided
	hasDueDate := req.DueDate != nil
	hasNumberOfPayments := req.NumberOfPayments != nil && *req.NumberOfPayments > 0
	if !hasDueDate && !hasNumberOfPayments {
		return fmt.Errorf("either due_date or number_of_payments must be provided")
	}
	return nil
}

func (s *debtService) validateUpdateDebtListRequest(req *entities.UpdateDebtListRequest) error {
	if req.TotalAmount != nil && *req.TotalAmount == "" {
		return entities.ErrInvalidAmount
	}
	if req.Currency != nil && *req.Currency == "" {
		return entities.ErrInvalidCurrency
	}
	return nil
}

func (s *debtService) validateCreateDebtItemRequest(req *entities.CreateDebtItemRequest) error {
	if req.DebtListID == uuid.Nil {
		return entities.ErrInvalidInput
	}
	if req.Amount == "" {
		return entities.ErrInvalidAmount
	}
	if req.PaymentMethod == "" {
		return entities.ErrInvalidPaymentMethod
	}
	
	// Validate payment method is one of the allowed values
	validPaymentMethods := map[string]bool{
		"cash":           true,
		"bank_transfer":  true,
		"check":          true,
		"digital_wallet": true,
		"other":          true,
	}
	if !validPaymentMethods[req.PaymentMethod] {
		return entities.ErrInvalidPaymentMethod
	}
	
	return nil
}

func (s *debtService) validateUpdateDebtItemRequest(req *entities.UpdateDebtItemRequest) error {
	if req.Amount != nil && *req.Amount == "" {
		return entities.ErrInvalidAmount
	}
	if req.Currency != nil && *req.Currency == "" {
		return entities.ErrInvalidCurrency
	}
	if req.PaymentMethod != nil && *req.PaymentMethod == "" {
		return entities.ErrInvalidPaymentMethod
	}
	return nil
}






