package services

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"

	"pay-your-dues/internal/domain/entities"
	"pay-your-dues/internal/domain/interfaces"
)

// authService implements the AuthService interface
type authService struct {
	userRepo          interfaces.UserRepository
	contactService    interfaces.ContactService
	userSettingsService interfaces.UserSettingsService
	jwtSecret         string
	jwtExpiry         time.Duration
}

// NewAuthService creates a new auth service
func NewAuthService(
	userRepo interfaces.UserRepository,
	contactService interfaces.ContactService,
	userSettingsService interfaces.UserSettingsService,
	jwtSecret string,
	jwtExpiry string,
) (interfaces.AuthService, error) {
	duration, err := time.ParseDuration(jwtExpiry)
	if err != nil {
		return nil, fmt.Errorf("invalid JWT expiry duration: %w", err)
	}

	return &authService{
		userRepo:            userRepo,
		contactService:      contactService,
		userSettingsService: userSettingsService,
		jwtSecret:           jwtSecret,
		jwtExpiry:           duration,
	}, nil
}

func (s *authService) Register(ctx context.Context, req *entities.CreateUserRequest) (*entities.RegisterResponse, error) {
	// Validate input
	if err := s.validateCreateUserRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check if user already exists
	exists, err := s.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check if user exists: %w", err)
	}
	if exists {
		return nil, entities.ErrUserAlreadyExists
	}

	// Check if phone number already exists (if provided)
	if req.Phone != nil && *req.Phone != "" {
		phoneExists, err := s.userRepo.ExistsByPhone(ctx, *req.Phone)
		if err != nil {
			return nil, fmt.Errorf("failed to check if phone exists: %w", err)
		}
		if phoneExists {
			return nil, entities.ErrPhoneNumberExists
		}
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user entity
	user := &entities.User{
		ID:           uuid.New(),
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Phone:        req.Phone,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Validate user entity
	if err := user.IsValid(); err != nil {
		return nil, fmt.Errorf("invalid user entity: %w", err)
	}

	// Create user in repository
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Create contacts for the new user based on existing contacts that have their email
	if err := s.contactService.CreateContactsForNewUser(ctx, user.ID, user.Email); err != nil {
		// Log the error but don't fail registration
		logger := zerolog.Ctx(ctx)
		logger.Warn().
			Err(err).
			Str("user_id", user.ID.String()).
			Str("user_email", user.Email).
			Msg("Failed to create contacts for new user during registration")
	}

	// Create default user settings for the new user
	if s.userSettingsService != nil {
		if _, err := s.userSettingsService.GetOrCreateUserSettings(ctx, user.ID); err != nil {
			// Log the error but don't fail registration
			logger := zerolog.Ctx(ctx)
			logger.Warn().
				Err(err).
				Str("user_id", user.ID.String()).
				Str("user_email", user.Email).
				Msg("Failed to create default user settings during registration")
		}
	}

	return &entities.RegisterResponse{
		User: *user,
	}, nil
}

func (s *authService) Login(ctx context.Context, req *entities.LoginRequest) (*entities.LoginResponse, error) {
	// Validate input
	if err := s.validateLoginRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if err == entities.ErrUserNotFound {
			return nil, entities.ErrInvalidCredentials
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, entities.ErrInvalidCredentials
	}

	// Generate JWT token
	token, err := s.GenerateJWT(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate JWT token: %w", err)
	}

	return &entities.LoginResponse{
		Token: token,
		User:  *user,
	}, nil
}

func (s *authService) ValidateToken(ctx context.Context, tokenString string) (uuid.UUID, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userIDStr, ok := claims["user_id"].(string)
		if !ok {
			return uuid.Nil, entities.ErrInvalidToken
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return uuid.Nil, fmt.Errorf("invalid user ID in token: %w", err)
		}

		// Check if token is expired
		if exp, ok := claims["exp"].(float64); ok {
			if time.Unix(int64(exp), 0).Before(time.Now()) {
				return uuid.Nil, entities.ErrTokenExpired
			}
		}

		return userID, nil
	}

	return uuid.Nil, entities.ErrInvalidToken
}

func (s *authService) GenerateJWT(ctx context.Context, userID uuid.UUID) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"exp":     time.Now().Add(s.jwtExpiry).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

func (s *authService) validateCreateUserRequest(req *entities.CreateUserRequest) error {
	if req.Email == "" {
		return entities.ErrInvalidEmail
	}
	if req.Password == "" || len(req.Password) < 6 {
		return entities.ErrInvalidPassword
	}
	if req.FirstName == "" {
		return entities.ErrInvalidFirstName
	}
	if req.LastName == "" {
		return entities.ErrInvalidLastName
	}
	return nil
}

func (s *authService) validateLoginRequest(req *entities.LoginRequest) error {
	if req.Email == "" {
		return entities.ErrInvalidEmail
	}
	if req.Password == "" {
		return entities.ErrInvalidPassword
	}
	return nil
}
