package notification

import (
	"context"
	"fmt"

	"pay-your-dues/internal/domain/interfaces"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// ContactInfo contains all contact and debt information needed for notifications
type ContactInfo struct {
	// Contact details
	ContactID   uuid.UUID
	ContactName string
	Email       *string
	Phone       *string

	// User details
	UserID        uuid.UUID
	UserFirstName string
	UserLastName  string
	UserEmail     string

	// Debt details
	DebtListID        uuid.UUID
	TotalDebt         decimal.Decimal
	RemainingDebt     decimal.Decimal
	PaidAmount        decimal.Decimal
	Currency          string
	DebtType          string
	InstallmentPlan   *string
	NumberOfPayments  *int
	InstallmentAmount *decimal.Decimal
}

// ContactFetcher handles fetching contact and debt information for notifications
type ContactFetcher struct {
	db              *gorm.DB
	debtListRepo    interfaces.DebtListRepository
	contactRepo     interfaces.ContactRepository
	userRepo        interfaces.UserRepository
}

// NewContactFetcher creates a new contact fetcher instance
func NewContactFetcher(
	db *gorm.DB,
	debtListRepo interfaces.DebtListRepository,
	contactRepo interfaces.ContactRepository,
	userRepo interfaces.UserRepository,
) *ContactFetcher {
	return &ContactFetcher{
		db:           db,
		debtListRepo: debtListRepo,
		contactRepo:  contactRepo,
		userRepo:     userRepo,
	}
}

// GetContactInfoForDebtList fetches complete contact and debt information
func (cf *ContactFetcher) GetContactInfoForDebtList(debtListID uuid.UUID) (*ContactInfo, error) {
	ctx := context.Background()

	// Get debt list using repository
	debtList, err := cf.debtListRepo.GetByID(ctx, debtListID)
	if err != nil {
		return nil, fmt.Errorf("failed to get debt list: %w", err)
	}

	// Get user using repository
	user, err := cf.userRepo.GetByID(ctx, debtList.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Get contact using repository
	contact, err := cf.contactRepo.GetByID(ctx, debtList.ContactID)
	if err != nil {
		return nil, fmt.Errorf("failed to get contact: %w", err)
	}

	// Get user-contact relation for email/phone
	userContact, err := cf.contactRepo.GetUserContactRelation(ctx, debtList.UserID, debtList.ContactID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user-contact relation: %w", err)
	}

	// Build ContactInfo from repository data
	info := &ContactInfo{
		ContactID:         contact.ID,
		ContactName:       userContact.Name,
		Email:             userContact.Email,
		Phone:             userContact.Phone,
		UserID:            user.ID,
		UserFirstName:     user.FirstName,
		UserLastName:      user.LastName,
		UserEmail:         user.Email,
		DebtListID:        debtList.ID,
		TotalDebt:         debtList.TotalAmount,
		RemainingDebt:     debtList.TotalRemainingDebt,
		PaidAmount:        debtList.TotalPaymentsMade,
		Currency:          debtList.Currency,
		DebtType:          debtList.DebtType,
		InstallmentPlan:   &debtList.InstallmentPlan,
		NumberOfPayments:  debtList.NumberOfPayments,
		InstallmentAmount: &debtList.InstallmentAmount,
	}

	return info, nil
}

// GetContactInfoByRecipientType gets contact info based on recipient type (user or contact)
func (cf *ContactFetcher) GetContactInfoByRecipientType(debtListID uuid.UUID, recipientType string) (*RecipientInfo, error) {
	contactInfo, err := cf.GetContactInfoForDebtList(debtListID)
	if err != nil {
		return nil, err
	}

	recipient := &RecipientInfo{
		DebtListID:    contactInfo.DebtListID,
		TotalDebt:     contactInfo.TotalDebt,
		RemainingDebt: contactInfo.RemainingDebt,
		PaidAmount:    contactInfo.PaidAmount,
		Currency:      contactInfo.Currency,
		DebtType:      contactInfo.DebtType,
	}

	if recipientType == "user" {
		// Notification goes to the user (debt owner)
		recipient.RecipientID = contactInfo.UserID
		recipient.RecipientName = fmt.Sprintf("%s %s", contactInfo.UserFirstName, contactInfo.UserLastName)
		recipient.RecipientEmail = &contactInfo.UserEmail
		recipient.OtherPartyName = contactInfo.ContactName
	} else {
		// Notification goes to the contact
		recipient.RecipientID = contactInfo.ContactID
		recipient.RecipientName = contactInfo.ContactName
		recipient.RecipientEmail = contactInfo.Email
		recipient.RecipientPhone = contactInfo.Phone
		recipient.OtherPartyName = fmt.Sprintf("%s %s", contactInfo.UserFirstName, contactInfo.UserLastName)
	}

	return recipient, nil
}

// RecipientInfo contains recipient-specific information
type RecipientInfo struct {
	RecipientID    uuid.UUID
	RecipientName  string
	RecipientEmail *string
	RecipientPhone *string
	OtherPartyName string

	// Debt info
	DebtListID    uuid.UUID
	TotalDebt     decimal.Decimal
	RemainingDebt decimal.Decimal
	PaidAmount    decimal.Decimal
	Currency      string
	DebtType      string
}

// GetUserSettings fetches user settings for notification preferences
func (cf *ContactFetcher) GetUserSettings(userID uuid.UUID) (*UserNotificationSettings, error) {
	var settings UserNotificationSettings

	// Use GORM's First method instead of raw SQL
	err := cf.db.Where("user_id = ?", userID).First(&settings).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Return default settings if not found
			return &UserNotificationSettings{
				UserID:                    userID,
				NotificationEmail:         true,
				NotificationSMS:           false,
				NotificationWebhook:       false,
				NotificationReminderDays:  []int64{7, 3, 1},
				NotificationTime:          "09:00:00",
				OverdueReminderFrequency:  "daily",
				EventNotificationsEnabled: true,
				NotifyContactOnPayment:    true,
			}, nil
		}
		return nil, fmt.Errorf("failed to fetch user settings: %w", err)
	}

	return &settings, nil
}

// UserNotificationSettings contains user notification preferences
type UserNotificationSettings struct {
	ID                        uuid.UUID
	UserID                    uuid.UUID
	NotificationEmail         bool
	NotificationSMS           bool
	NotificationWebhook       bool
	NotificationReminderDays  []int64
	NotificationTime          string
	OverdueReminderFrequency  string
	CustomEmailMessage        *string
	CustomSMSMessage          *string
	SlackWebhookURL           *string
	TelegramBotToken          *string
	TelegramChatID            *string
	DiscordWebhookURL         *string
	EventNotificationsEnabled bool
	NotifyContactOnPayment    bool
}

// HasEmail checks if contact has an email address
func (ci *ContactInfo) HasEmail() bool {
	return ci.Email != nil && *ci.Email != ""
}

// HasPhone checks if contact has a phone number
func (ci *ContactInfo) HasPhone() bool {
	return ci.Phone != nil && *ci.Phone != ""
}

// GetFullUserName returns the user's full name
func (ci *ContactInfo) GetFullUserName() string {
	return fmt.Sprintf("%s %s", ci.UserFirstName, ci.UserLastName)
}

// HasEmail checks if recipient has an email address
func (ri *RecipientInfo) HasEmail() bool {
	return ri.RecipientEmail != nil && *ri.RecipientEmail != ""
}

// HasPhone checks if recipient has a phone number
func (ri *RecipientInfo) HasPhone() bool {
	return ri.RecipientPhone != nil && *ri.RecipientPhone != ""
}

