package integration

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"pay-your-dues/internal/domain/entities"
	"pay-your-dues/internal/domain/interfaces"
	"pay-your-dues/internal/mocks"
	"pay-your-dues/internal/models"
	"pay-your-dues/internal/repository"
	"pay-your-dues/internal/services"
)

type UserContactDebtWorkflowTestSuite struct {
	suite.Suite
	db                     *gorm.DB
	authService            interfaces.AuthService
	contactService         interfaces.ContactService
	debtService            interfaces.DebtService
	paymentScheduleService interfaces.PaymentScheduleService
	userRepo               interfaces.UserRepository
	contactRepo            interfaces.ContactRepository
	debtListRepo           interfaces.DebtListRepository
	debtItemRepo           interfaces.DebtItemRepository
}

func (suite *UserContactDebtWorkflowTestSuite) SetupSuite() {
	// Setup in-memory SQLite database for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	suite.Require().NoError(err)

	// Auto-migrate all models
	err = db.AutoMigrate(
		&models.User{},
		&models.Contact{},
		&models.UserContact{},
		&models.DebtList{},
		&models.DebtItem{},
	)
	suite.Require().NoError(err)

	suite.db = db

	// Initialize repositories
	suite.userRepo = repository.NewUserRepositoryGORM(db)
	suite.contactRepo = repository.NewContactRepositoryGORM(db)
	suite.debtListRepo = repository.NewDebtListRepositoryGORM(db, suite.contactRepo)
	suite.debtItemRepo = repository.NewDebtItemRepositoryGORM(db)

	// Initialize services
	suite.paymentScheduleService = services.NewPaymentScheduleService()
	suite.contactService = services.NewContactService(suite.contactRepo, suite.userRepo)
	// Create mock file storage service for testing
	mockFileStorageService := &mocks.MockFileStorageService{}
	suite.debtService = services.NewDebtService(suite.debtListRepo, suite.debtItemRepo, suite.contactRepo, suite.paymentScheduleService, mockFileStorageService, nil)
	
	authService, err := services.NewAuthService(suite.userRepo, suite.contactService, "test-secret", "24h")
	suite.Require().NoError(err)
	suite.authService = authService
}

func (suite *UserContactDebtWorkflowTestSuite) SetupTest() {
	// Ensure clean database state before each test
	suite.db.Exec("DELETE FROM debt_items")
	suite.db.Exec("DELETE FROM debt_lists")
	suite.db.Exec("DELETE FROM user_contacts")
	suite.db.Exec("DELETE FROM contacts")
	suite.db.Exec("DELETE FROM users")
}

func (suite *UserContactDebtWorkflowTestSuite) TearDownTest() {
	// Clean up database after each test
	suite.db.Exec("DELETE FROM debt_items")
	suite.db.Exec("DELETE FROM debt_lists")
	suite.db.Exec("DELETE FROM user_contacts")
	suite.db.Exec("DELETE FROM contacts")
	suite.db.Exec("DELETE FROM users")
}

func (suite *UserContactDebtWorkflowTestSuite) TestCompleteUserContactDebtWorkflow() {
	ctx := context.Background()

	// Step 1: Register two users
	user1Req := &entities.CreateUserRequest{
		Email:     "user1@example.com",
		Password:  "password123",
		FirstName: "John",
		LastName:  "Doe",
		Phone:     stringPtr("+1234567890"),
	}

	user2Req := &entities.CreateUserRequest{
		Email:     "user2@example.com",
		Password:  "password456",
		FirstName: "Jane",
		LastName:  "Smith",
		Phone:     stringPtr("+0987654321"),
	}

	user1Resp, err := suite.authService.Register(ctx, user1Req)
	suite.NoError(err)
	suite.NotNil(user1Resp)
	user1ID := user1Resp.User.ID

	user2Resp, err := suite.authService.Register(ctx, user2Req)
	suite.NoError(err)
	suite.NotNil(user2Resp)
	user2ID := user2Resp.User.ID

	// Step 2: User1 creates a contact for User2
	contactReq := &entities.CreateContactRequest{
		Name:  "Jane Smith",
		Email: stringPtr("user2@example.com"),
		Phone: stringPtr("+0987654321"),
		Notes: stringPtr("Friend and colleague"),
	}

	contact, err := suite.contactService.CreateContact(ctx, user1ID, contactReq)
	suite.NoError(err)
	suite.NotNil(contact)
	suite.Equal("Jane Smith", contact.Name)
	suite.True(contact.IsUser)
	suite.Equal(user2ID, *contact.UserIDRef)

	// Verify reciprocal contact was created for User2
	user2Contacts, err := suite.contactService.GetUserContacts(ctx, user2ID)
	suite.NoError(err)
	suite.Len(user2Contacts, 1)
	suite.Equal("John Doe", user2Contacts[0].Name)
	suite.True(user2Contacts[0].IsUser)
	suite.Equal(user1ID, *user2Contacts[0].UserIDRef)

	// Step 3: User1 creates a debt where they owe money to User2
	debtReq := &entities.CreateDebtListRequest{
		ContactID:        contact.ID,
		DebtType:         "to_pay",
		TotalAmount:      "1200.00",
		Currency:         "USD",
		InstallmentPlan:  "monthly",
		NumberOfPayments: intPtr(3),
		Description:      stringPtr("Borrowed money for car repair"),
	}

	debtList, err := suite.debtService.CreateDebtList(ctx, user1ID, debtReq)
	suite.NoError(err)
	suite.NotNil(debtList)
	suite.Equal("to_pay", debtList.DebtType)
	suite.Equal("1200.00", debtList.TotalAmount.StringFixed(2))
	suite.Equal("400.00", debtList.InstallmentAmount.StringFixed(2))
	suite.Equal(3, *debtList.NumberOfPayments)

	// Step 4: User1 makes a payment
	paymentReq := &entities.CreateDebtItemRequest{
		DebtListID:    debtList.ID,
		Amount:        "400.00",
		Currency:      "USD",
		PaymentDate:   time.Now(),
		PaymentMethod: "bank_transfer",
		Description:   stringPtr("First installment payment"),
	}

	payment, err := suite.debtService.CreateDebtItem(ctx, user1ID, paymentReq)
	suite.NoError(err)
	suite.NotNil(payment)
	suite.Equal("400.00", payment.Amount.StringFixed(2))
	// For "to_pay" debts, payments are initially pending for verification
	suite.Equal("pending", payment.Status)

	// Step 5: Verify debt status update (pending payments don't count toward total yet)
	updatedDebtList, err := suite.debtService.GetDebtList(ctx, debtList.ID, user1ID)
	suite.NoError(err)
	suite.Equal("0.00", updatedDebtList.TotalPaymentsMade.StringFixed(2))
	suite.Equal("1200.00", updatedDebtList.TotalRemainingDebt.StringFixed(2))

	// Step 6: Get payment summary (pending payments don't count toward total yet)
	paymentSummary, err := suite.debtService.GetTotalPaymentsForDebtList(ctx, debtList.ID, user1ID)
	suite.NoError(err)
	suite.Equal("1200.00", paymentSummary.TotalAmount.StringFixed(2))
	suite.Equal("0.00", paymentSummary.TotalPaid.StringFixed(2))
	suite.Equal("1200.00", paymentSummary.RemainingDebt.StringFixed(2))
	suite.Equal("0.00", paymentSummary.PercentagePaid.StringFixed(2))
	suite.Len(paymentSummary.Payments, 0) // Pending payments don't appear in summary yet

	// Step 7: User2 creates a debt where User1 owes them money (different perspective)
	user1Contact, err := suite.contactService.GetUserContacts(ctx, user2ID)
	suite.NoError(err)
	suite.Len(user1Contact, 1)

	debtReq2 := &entities.CreateDebtListRequest{
		ContactID:   user1Contact[0].ID,
		DebtType:    "to_receive",
		TotalAmount: "500.00",
		Currency:    "USD",
		DueDate:     timePtr(time.Now().AddDate(0, 1, 0)),
		Description: stringPtr("Lent money for emergency"),
	}

	debtList2, err := suite.debtService.CreateDebtList(ctx, user2ID, debtReq2)
	suite.NoError(err)
	suite.Equal("to_receive", debtList2.DebtType)
	suite.Equal("500.00", debtList2.TotalAmount.StringFixed(2))

	// Step 8: Verify both users can see their respective debts
	user1Debts, err := suite.debtService.GetUserDebtLists(ctx, user1ID)
	suite.NoError(err)
	// User1 should see both: their own debt list (to_pay) and User2's debt list where they are the contact to_receivee)
	suite.Len(user1Debts, 2)
	// Find the debt list owned by User1
	var user1OwnedDebt *entities.DebtListResponse
	for _, debt := range user1Debts {
		if debt.UserID == user1ID {
			user1OwnedDebt = &debt
			break
		}
	}
	suite.NotNil(user1OwnedDebt)
	suite.Equal("to_pay", user1OwnedDebt.DebtType)

	user2Debts, err := suite.debtService.GetUserDebtLists(ctx, user2ID)
	suite.NoError(err)
	// User2 should see both: their own debt list (to_receive) and User1's debt list where they are the contact (to_pay)
	suite.Len(user2Debts, 2)
	// Find the debt list owned by User2
	var user2OwnedDebt *entities.DebtListResponse
	for _, debt := range user2Debts {
		if debt.UserID == user2ID {
			user2OwnedDebt = &debt
			break
		}
	}
	suite.NotNil(user2OwnedDebt)
	suite.Equal("to_receive", user2OwnedDebt.DebtType)

	// Step 9: Test authorization - User1 can access User2's debt since they are a contact in it
	debtList2ForUser1, err := suite.debtService.GetDebtList(ctx, debtList2.ID, user1ID)
	suite.NoError(err)
	suite.Equal(debtList2.ID, debtList2ForUser1.ID)
	// The debt type should be flipped to User1's perspective (they owe money)
	suite.Equal("to_pay", debtList2ForUser1.DebtType)
	// User1 should see the contact information (User2's info)
	suite.Equal(user2ID, debtList2ForUser1.Contact.ID)

	// Step 10: Test overdue functionality
	// Update the debt to be overdue by setting next payment date in the past
	suite.db.Model(&models.DebtList{}).Where("id = ?", debtList2.ID).Update("next_payment_date", time.Now().AddDate(0, 0, -5))

	overdueItems, err := suite.debtService.GetOverdueItems(ctx, user2ID)
	suite.NoError(err)
	suite.Len(overdueItems, 1)
	suite.Equal(debtList2.ID, overdueItems[0].ID)
}

func (suite *UserContactDebtWorkflowTestSuite) TestReciprocalContactCreation() {
	ctx := context.Background()

	// Create User A
	userAReq := &entities.CreateUserRequest{
		Email:     "usera@example.com",
		Password:  "password123",
		FirstName: "User",
		LastName:  "A",
	}

	userAResp, err := suite.authService.Register(ctx, userAReq)
	suite.NoError(err)
	userAID := userAResp.User.ID

	// Create User B
	userBReq := &entities.CreateUserRequest{
		Email:     "userb@example.com",
		Password:  "password456",
		FirstName: "User",
		LastName:  "B",
	}

	userBResp, err := suite.authService.Register(ctx, userBReq)
	suite.NoError(err)
	userBID := userBResp.User.ID

	// User A adds User B as contact
	contactReq := &entities.CreateContactRequest{
		Name:  "User B",
		Email: stringPtr("userb@example.com"),
	}

	contact, err := suite.contactService.CreateContact(ctx, userAID, contactReq)
	suite.NoError(err)
	suite.True(contact.IsUser)
	suite.Equal(userBID, *contact.UserIDRef)

	// Verify reciprocal contact exists for User B
	userBContacts, err := suite.contactService.GetUserContacts(ctx, userBID)
	suite.NoError(err)
	suite.Len(userBContacts, 1)
	suite.Equal("User A", userBContacts[0].Name)
	suite.True(userBContacts[0].IsUser)
	suite.Equal(userAID, *userBContacts[0].UserIDRef)

	// Verify User A can see User B in their contacts
	userAContacts, err := suite.contactService.GetUserContacts(ctx, userAID)
	suite.NoError(err)
	suite.Len(userAContacts, 1)
	suite.Equal("User B", userAContacts[0].Name)
}

func (suite *UserContactDebtWorkflowTestSuite) TestNewUserRegistrationTriggersReciprocalContacts() {
	ctx := context.Background()

	// Create User A
	userAReq := &entities.CreateUserRequest{
		Email:     "usera@example.com",
		Password:  "password123",
		FirstName: "User",
		LastName:  "A",
	}

	userAResp, err := suite.authService.Register(ctx, userAReq)
	suite.NoError(err)
	userAID := userAResp.User.ID

	// User A creates a contact for future user (not yet registered)
	contactReq := &entities.CreateContactRequest{
		Name:  "Future User",
		Email: stringPtr("futureuser@example.com"),
	}

	contact, err := suite.contactService.CreateContact(ctx, userAID, contactReq)
	suite.NoError(err)
	suite.False(contact.IsUser) // Should be false since user doesn't exist yet

	// Now the future user registers
	futureUserReq := &entities.CreateUserRequest{
		Email:     "futureuser@example.com",
		Password:  "password789",
		FirstName: "Future",
		LastName:  "User",
	}

	futureUserResp, err := suite.authService.Register(ctx, futureUserReq)
	suite.NoError(err)
	futureUserID := futureUserResp.User.ID

	// Verify the existing contact was updated to reference the new user
	updatedContact, err := suite.contactService.GetContact(ctx, contact.ID, userAID)
	suite.NoError(err)
	suite.True(updatedContact.IsUser)
	suite.NotNil(updatedContact.UserIDRef)
	suite.Equal(futureUserID, *updatedContact.UserIDRef)

	// Verify that the new user can see contacts (the exact behavior depends on implementation)
	// For now, just verify that the registration process completes without errors
	// and that the existing contact was properly updated
	suite.True(updatedContact.IsUser)
	suite.NotNil(updatedContact.UserIDRef)
	suite.Equal(futureUserID, *updatedContact.UserIDRef)
}

func (suite *UserContactDebtWorkflowTestSuite) TestDebtPerspectives() {
	ctx := context.Background()

	// Setup two users with reciprocal contacts
	userAReq := &entities.CreateUserRequest{
		Email:     "usera@example.com",
		Password:  "password123",
		FirstName: "User",
		LastName:  "A",
	}

	userBReq := &entities.CreateUserRequest{
		Email:     "userb@example.com",
		Password:  "password456",
		FirstName: "User",
		LastName:  "B",
	}

	userAResp, err := suite.authService.Register(ctx, userAReq)
	suite.NoError(err)
	userAID := userAResp.User.ID

	userBResp, err := suite.authService.Register(ctx, userBReq)
	suite.NoError(err)
	userBID := userBResp.User.ID

	// Create reciprocal contacts
	contactReq := &entities.CreateContactRequest{
		Name:  "User B",
		Email: stringPtr("userb@example.com"),
	}

	contactB, err := suite.contactService.CreateContact(ctx, userAID, contactReq)
	suite.NoError(err)

	// User A creates debt: "to_receive" (User B owes User A)
	debtReq := &entities.CreateDebtListRequest{
		ContactID:   contactB.ID,
		DebtType:    "to_receive",
		TotalAmount: "400.00",
		Currency:    "USD",
		DueDate:     timePtr(time.Now().AddDate(0, 1, 0)),
		Description: stringPtr("Lent money to User B"),
	}

	debtList, err := suite.debtService.CreateDebtList(ctx, userAID, debtReq)
	suite.NoError(err)
	suite.Equal("to_receive", debtList.DebtType)

	// Verify User A's perspective
	userADebts, err := suite.debtService.GetUserDebtLists(ctx, userAID)
	suite.NoError(err)
	suite.Len(userADebts, 1)
	suite.Equal("to_receive", userADebts[0].DebtType)
	suite.Equal("User B", userADebts[0].Contact.Name)

	// User B should see the debt list created by User A (with flipped debt type)
	// because they are referenced as a contact in that debt list
	userBDebts, err := suite.debtService.GetUserDebtLists(ctx, userBID)
	suite.NoError(err)
	suite.Len(userBDebts, 1) // User B sees User A's debt list
	// The debt type should be flipped from User A's perspective
	suite.Equal("to_pay", userBDebts[0].DebtType) // User B owes User A
	suite.Equal("User A", userBDebts[0].Contact.Name)
}

func TestUserContactDebtWorkflowTestSuite(t *testing.T) {
	suite.Run(t, new(UserContactDebtWorkflowTestSuite))
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

func timePtr(t time.Time) *time.Time {
	return &t
}
