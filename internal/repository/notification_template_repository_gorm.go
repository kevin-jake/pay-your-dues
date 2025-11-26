package repository

import (
	"pay-your-dues/internal/domain/interfaces"
	"pay-your-dues/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// NotificationTemplateRepositoryGORM implements the NotificationTemplateRepository interface using GORM
type NotificationTemplateRepositoryGORM struct {
	db *gorm.DB
}

// NewNotificationTemplateRepositoryGORM creates a new notification template repository instance
func NewNotificationTemplateRepositoryGORM(db *gorm.DB) interfaces.NotificationTemplateRepository {
	return &NotificationTemplateRepositoryGORM{db: db}
}

// Create creates a new notification template
func (r *NotificationTemplateRepositoryGORM) Create(template *models.NotificationTemplate) error {
	if template.ID == uuid.Nil {
		template.ID = uuid.New()
	}
	return r.db.Create(template).Error
}

// GetByID retrieves a notification template by ID
func (r *NotificationTemplateRepositoryGORM) GetByID(id uuid.UUID) (*models.NotificationTemplate, error) {
	var template models.NotificationTemplate
	err := r.db.Where("id = ?", id).First(&template).Error
	if err != nil {
		return nil, err
	}
	return &template, nil
}

// Update updates a notification template
func (r *NotificationTemplateRepositoryGORM) Update(template *models.NotificationTemplate) error {
	return r.db.Save(template).Error
}

// Delete deletes a notification template
func (r *NotificationTemplateRepositoryGORM) Delete(id uuid.UUID) error {
	return r.db.Where("id = ?", id).Delete(&models.NotificationTemplate{}).Error
}

// GetByUserID retrieves all templates for a user
func (r *NotificationTemplateRepositoryGORM) GetByUserID(userID uuid.UUID) ([]*models.NotificationTemplate, error) {
	var templates []*models.NotificationTemplate
	err := r.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&templates).Error
	if err != nil {
		return nil, err
	}
	return templates, nil
}

// GetByType retrieves templates by type
func (r *NotificationTemplateRepositoryGORM) GetByType(templateType string) ([]*models.NotificationTemplate, error) {
	var templates []*models.NotificationTemplate
	err := r.db.Where("template_type = ?", templateType).
		Order("created_at DESC").
		Find(&templates).Error
	if err != nil {
		return nil, err
	}
	return templates, nil
}

// GetDefaultTemplates retrieves all default (system) templates
func (r *NotificationTemplateRepositoryGORM) GetDefaultTemplates() ([]*models.NotificationTemplate, error) {
	var templates []*models.NotificationTemplate
	err := r.db.Where("is_default = ? AND user_id IS NULL", true).
		Order("template_type, created_at DESC").
		Find(&templates).Error
	if err != nil {
		return nil, err
	}
	return templates, nil
}

// GetUserTemplate retrieves a user's template for a specific type, falling back to default
func (r *NotificationTemplateRepositoryGORM) GetUserTemplate(userID uuid.UUID, templateType string) (*models.NotificationTemplate, error) {
	var template models.NotificationTemplate
	
	// First try to get user's custom template
	err := r.db.Where("user_id = ? AND template_type = ?", userID, templateType).
		Order("is_default DESC, created_at DESC").
		First(&template).Error
	
	if err == gorm.ErrRecordNotFound {
		// Fall back to default template
		err = r.db.Where("user_id IS NULL AND template_type = ? AND is_default = ?",
			templateType, true).
			First(&template).Error
	}
	
	if err != nil {
		return nil, err
	}
	
	return &template, nil
}

// SetAsDefault sets a template as the default for the user
func (r *NotificationTemplateRepositoryGORM) SetAsDefault(id uuid.UUID, userID uuid.UUID) error {
	// Get the template to find its type
	var template models.NotificationTemplate
	if err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&template).Error; err != nil {
		return err
	}

	// Start a transaction
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Unset any other default templates of the same type for this user
		if err := tx.Model(&models.NotificationTemplate{}).
			Where("user_id = ? AND template_type = ? AND id != ?",
				userID, template.TemplateType, id).
			Update("is_default", false).Error; err != nil {
			return err
		}

		// Set this template as default
		if err := tx.Model(&models.NotificationTemplate{}).
			Where("id = ?", id).
			Update("is_default", true).Error; err != nil {
			return err
		}

		return nil
	})
}

// UnsetDefault unsets the default template for a user and type
func (r *NotificationTemplateRepositoryGORM) UnsetDefault(userID uuid.UUID, templateType string) error {
	return r.db.Model(&models.NotificationTemplate{}).
		Where("user_id = ? AND template_type = ?", userID, templateType).
		Update("is_default", false).Error
}

// GetUserTemplatesByType retrieves all templates of a specific type for a user
func (r *NotificationTemplateRepositoryGORM) GetUserTemplatesByType(userID uuid.UUID, templateType string) ([]*models.NotificationTemplate, error) {
	var templates []*models.NotificationTemplate
	err := r.db.Where("user_id = ? AND template_type = ?", userID, templateType).
		Order("is_default DESC, created_at DESC").
		Find(&templates).Error
	if err != nil {
		return nil, err
	}
	return templates, nil
}

// BelongsToUser checks if a template belongs to a user
func (r *NotificationTemplateRepositoryGORM) BelongsToUser(templateID uuid.UUID, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.NotificationTemplate{}).
		Where("id = ? AND user_id = ?", templateID, userID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// CreateDefaultTemplates creates system default templates
func (r *NotificationTemplateRepositoryGORM) CreateDefaultTemplates() error {
	defaultTemplates := []*models.NotificationTemplate{
		{
			ID:           uuid.New(),
			UserID:       nil, // System template
			TemplateName: "Default Payment Reminder",
			TemplateType: "email",
			Subject:      strPtr("Payment Reminder"),
			Body:         getDefaultEmailBody(),
			IsDefault:    true,
			Variables:    []string{"user_first_name", "contact_name", "amount", "currency", "due_date", "days_until_due"},
		},
		{
			ID:           uuid.New(),
			UserID:       nil,
			TemplateName: "Default SMS Reminder",
			TemplateType: "sms",
			Body:         "Reminder: Payment of {{currency}} {{amount}} due to {{contact_name}} on {{due_date}} ({{days_until_due}} days).",
			IsDefault:    true,
			Variables:    []string{"currency", "amount", "contact_name", "due_date", "days_until_due"},
		},
	}

	for _, template := range defaultTemplates {
		// Check if template already exists
		var count int64
		err := r.db.Model(&models.NotificationTemplate{}).
			Where("user_id IS NULL AND template_type = ? AND is_default = ?",
				template.TemplateType, true).
			Count(&count).Error
		if err != nil {
			return err
		}

		// Only create if doesn't exist
		if count == 0 {
			if err := r.db.Create(template).Error; err != nil {
				return err
			}
		}
	}

	return nil
}

// Helper function for string pointer
func strPtr(s string) *string {
	return &s
}

// Helper function to get default email body
func getDefaultEmailBody() string {
	return `Hi {{user_first_name}},

This is a friendly reminder that a payment is due to {{contact_name}}.

Payment Details:
- Amount: {{currency}} {{amount}}
- Due Date: {{due_date}}
- Days Until Due: {{days_until_due}} days

Please make sure to complete this payment on time.

Best regards,
Pay Your Dues Team`
}

