package unit

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"

	"pay-your-dues/internal/domain/entities"
	"pay-your-dues/internal/domain/interfaces"
	"pay-your-dues/internal/mocks"
	"pay-your-dues/internal/services"
)

func TestAuthService_Register(t *testing.T) {
	tests := []struct {
		name           string
		request        *entities.CreateUserRequest
		setupMocks     func(*mocks.MockUserRepository, *mocks.MockContactService)
		expectedError  error
		expectSuccess  bool
		validateResult func(*testing.T, *entities.RegisterResponse)
	}{
		{
			name: "successful registration",
			request: &entities.CreateUserRequest{
				Email:     "test@example.com",
				Password:  "password123",
				FirstName: "John",
				LastName:  "Doe",
				Phone:     stringPtr("+1234567890"),
			},
			setupMocks: func(userRepo *mocks.MockUserRepository, contactService *mocks.MockContactService) {
				userRepo.On("ExistsByEmail", mock.Anything, "test@example.com").Return(false, nil)
				userRepo.On("ExistsByPhone", mock.Anything, "+1234567890").Return(false, nil)
				userRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.User")).Return(nil)
				contactService.On("CreateContactsForNewUser", mock.Anything, mock.AnythingOfType("uuid.UUID"), "test@example.com").Return(nil)
			},
			expectedError: nil,
			expectSuccess: true,
			validateResult: func(t *testing.T, resp *entities.RegisterResponse) {
				assert.Equal(t, "test@example.com", resp.User.Email)
				assert.Equal(t, "John", resp.User.FirstName)
				assert.Equal(t, "Doe", resp.User.LastName)
				assert.NotEmpty(t, resp.User.ID)
				assert.NotEmpty(t, resp.User.PasswordHash)
			},
		},
		{
			name: "user already exists",
			request: &entities.CreateUserRequest{
				Email:     "existing@example.com",
				Password:  "password123",
				FirstName: "Jane",
				LastName:  "Smith",
			},
			setupMocks: func(userRepo *mocks.MockUserRepository, contactService *mocks.MockContactService) {
				userRepo.On("ExistsByEmail", mock.Anything, "existing@example.com").Return(true, nil)
			},
			expectedError: entities.ErrUserAlreadyExists,
			expectSuccess: false,
		},
		{
			name: "invalid email",
			request: &entities.CreateUserRequest{
				Email:     "",
				Password:  "password123",
				FirstName: "John",
				LastName:  "Doe",
			},
			setupMocks:    func(userRepo *mocks.MockUserRepository, contactService *mocks.MockContactService) {},
			expectedError: entities.ErrInvalidEmail,
			expectSuccess: false,
		},
		{
			name: "password too short",
			request: &entities.CreateUserRequest{
				Email:     "test@example.com",
				Password:  "123",
				FirstName: "John",
				LastName:  "Doe",
			},
			setupMocks:    func(userRepo *mocks.MockUserRepository, contactService *mocks.MockContactService) {},
			expectedError: entities.ErrInvalidPassword,
			expectSuccess: false,
		},
		{
			name: "missing first name",
			request: &entities.CreateUserRequest{
				Email:     "test@example.com",
				Password:  "password123",
				FirstName: "",
				LastName:  "Doe",
			},
			setupMocks:    func(userRepo *mocks.MockUserRepository, contactService *mocks.MockContactService) {},
			expectedError: entities.ErrInvalidFirstName,
			expectSuccess: false,
		},
		{
			name: "phone number already exists",
			request: &entities.CreateUserRequest{
				Email:     "newuser@example.com",
				Password:  "password123",
				FirstName: "Jane",
				LastName:  "Smith",
				Phone:     stringPtr("+1234567890"),
			},
			setupMocks: func(userRepo *mocks.MockUserRepository, contactService *mocks.MockContactService) {
				userRepo.On("ExistsByEmail", mock.Anything, "newuser@example.com").Return(false, nil)
				userRepo.On("ExistsByPhone", mock.Anything, "+1234567890").Return(true, nil)
			},
			expectedError: entities.ErrPhoneNumberExists,
			expectSuccess: false,
		},
		{
			name: "successful registration without phone",
			request: &entities.CreateUserRequest{
				Email:     "nophone@example.com",
				Password:  "password123",
				FirstName: "Alice",
				LastName:  "Johnson",
				Phone:     nil,
			},
			setupMocks: func(userRepo *mocks.MockUserRepository, contactService *mocks.MockContactService) {
				userRepo.On("ExistsByEmail", mock.Anything, "nophone@example.com").Return(false, nil)
				userRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.User")).Return(nil)
				contactService.On("CreateContactsForNewUser", mock.Anything, mock.AnythingOfType("uuid.UUID"), "nophone@example.com").Return(nil)
			},
			expectedError: nil,
			expectSuccess: true,
			validateResult: func(t *testing.T, resp *entities.RegisterResponse) {
				assert.Equal(t, "nophone@example.com", resp.User.Email)
				assert.Equal(t, "Alice", resp.User.FirstName)
				assert.Equal(t, "Johnson", resp.User.LastName)
				assert.Nil(t, resp.User.Phone)
			},
		},
		{
			name: "successful registration with empty phone string",
			request: &entities.CreateUserRequest{
				Email:     "emptyphone@example.com",
				Password:  "password123",
				FirstName: "Bob",
				LastName:  "Williams",
				Phone:     stringPtr(""),
			},
			setupMocks: func(userRepo *mocks.MockUserRepository, contactService *mocks.MockContactService) {
				userRepo.On("ExistsByEmail", mock.Anything, "emptyphone@example.com").Return(false, nil)
				userRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.User")).Return(nil)
				contactService.On("CreateContactsForNewUser", mock.Anything, mock.AnythingOfType("uuid.UUID"), "emptyphone@example.com").Return(nil)
			},
			expectedError: nil,
			expectSuccess: true,
			validateResult: func(t *testing.T, resp *entities.RegisterResponse) {
				assert.Equal(t, "emptyphone@example.com", resp.User.Email)
				assert.Equal(t, "Bob", resp.User.FirstName)
				assert.Equal(t, "Williams", resp.User.LastName)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			userRepo := &mocks.MockUserRepository{}
			contactService := &mocks.MockContactService{}
			tt.setupMocks(userRepo, contactService)

			// Create service
			authService, err := services.NewAuthService(userRepo, contactService, nil, "test-secret", "24h")
			assert.NoError(t, err)

			// Execute
			ctx := context.Background()
			result, err := authService.Register(ctx, tt.request)

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
			userRepo.AssertExpectations(t)
			contactService.AssertExpectations(t)
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	// Setup test user
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	testUser := &entities.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: string(hashedPassword),
		FirstName:    "John",
		LastName:     "Doe",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	tests := []struct {
		name          string
		request       *entities.LoginRequest
		setupMocks    func(*mocks.MockUserRepository)
		expectedError error
		expectSuccess bool
	}{
		{
			name: "successful login",
			request: &entities.LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			setupMocks: func(userRepo *mocks.MockUserRepository) {
				userRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(testUser, nil)
			},
			expectedError: nil,
			expectSuccess: true,
		},
		{
			name: "invalid password",
			request: &entities.LoginRequest{
				Email:    "test@example.com",
				Password: "wrongpassword",
			},
			setupMocks: func(userRepo *mocks.MockUserRepository) {
				userRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(testUser, nil)
			},
			expectedError: entities.ErrInvalidCredentials,
			expectSuccess: false,
		},
		{
			name: "user not found",
			request: &entities.LoginRequest{
				Email:    "nonexistent@example.com",
				Password: "password123",
			},
			setupMocks: func(userRepo *mocks.MockUserRepository) {
				userRepo.On("GetByEmail", mock.Anything, "nonexistent@example.com").Return(nil, entities.ErrUserNotFound)
			},
			expectedError: entities.ErrInvalidCredentials,
			expectSuccess: false,
		},
		{
			name: "invalid email",
			request: &entities.LoginRequest{
				Email:    "",
				Password: "password123",
			},
			setupMocks:    func(userRepo *mocks.MockUserRepository) {},
			expectedError: entities.ErrInvalidEmail,
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			userRepo := &mocks.MockUserRepository{}
			contactService := &mocks.MockContactService{}
			tt.setupMocks(userRepo)

			// Create service
			authService, err := services.NewAuthService(userRepo, contactService, nil, "test-secret", "24h")
			assert.NoError(t, err)

			// Execute
			ctx := context.Background()
			result, err := authService.Login(ctx, tt.request)

			// Assert
			if tt.expectSuccess {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.NotEmpty(t, result.Token)
				assert.Equal(t, testUser.Email, result.User.Email)
			} else {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Nil(t, result)
			}

			// Verify mock expectations
			userRepo.AssertExpectations(t)
		})
	}
}

func TestAuthService_ValidateToken(t *testing.T) {
	tests := []struct {
		name          string
		tokenSetup    func(authService interfaces.AuthService) string
		expectedError error
		expectSuccess bool
	}{
		{
			name: "valid token",
			tokenSetup: func(authService interfaces.AuthService) string {
				userID := uuid.New()
				token, _ := authService.GenerateJWT(context.Background(), userID)
				return token
			},
			expectedError: nil,
			expectSuccess: true,
		},
		{
			name: "invalid token format",
			tokenSetup: func(authService interfaces.AuthService) string {
				return "invalid.token.format"
			},
			expectedError: entities.ErrInvalidToken,
			expectSuccess: false,
		},
		{
			name: "empty token",
			tokenSetup: func(authService interfaces.AuthService) string {
				return ""
			},
			expectedError: entities.ErrInvalidToken,
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			userRepo := &mocks.MockUserRepository{}
			contactService := &mocks.MockContactService{}

			authService, err := services.NewAuthService(userRepo, contactService, nil, "test-secret", "24h")
			assert.NoError(t, err)

			token := tt.tokenSetup(authService)

			// Execute
			ctx := context.Background()
			userID, err := authService.ValidateToken(ctx, token)

			// Assert
			if tt.expectSuccess {
				assert.NoError(t, err)
				assert.NotEqual(t, uuid.Nil, userID)
			} else {
				assert.Error(t, err)
				assert.Equal(t, uuid.Nil, userID)
			}
		})
	}
}
