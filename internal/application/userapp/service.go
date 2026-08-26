// Package userapp provides application-level use cases for users.
package userapp

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/joshua-sajeev/tessera/internal/domain/user"
	"github.com/joshua-sajeev/tessera/internal/ports"
)

// UserService manages the application-level logic for users.
type UserService struct {
	repo          ports.UserRepository
	apiKeyPrefix  string
	apiKeyVersion string
}

// NewUserService creates and returns a new UserService.
func NewUserService(repo ports.UserRepository, apiKeyPrefix, apiKeyVersion string) *UserService {
	return &UserService{
		repo:          repo,
		apiKeyPrefix:  apiKeyPrefix,
		apiKeyVersion: apiKeyVersion,
	}
}

// CreateUserInput contains the fields required to create a new user.
type CreateUserInput struct {
	Username     string
	Email        string
	StorageQuota int64
}

// UserDTO is a data transfer object representing a user's details.
type UserDTO struct {
	ID           uuid.UUID
	Username     string
	Email        string
	APIKey       string
	StorageQuota int64
	StorageUsed  int64
	Status       string
	CreatedAt    *time.Time
	UpdatedAt    *time.Time
}

// Create orchestrates user creation with all business logic
func (s *UserService) Create(ctx context.Context, input CreateUserInput) (*UserDTO, error) {
	if err := s.validateCreateInput(input); err != nil {
		return nil, err
	}

	generator := user.NewAPIKeyGenerator(s.apiKeyPrefix, s.apiKeyVersion)
	fullKey, keyID, err := generator.Generate()
	if err != nil {
		return nil, fmt.Errorf("failed to generate api key: %w", err)
	}

	hasher := user.NewKeyHasher()
	hash, err := hasher.Hash(fullKey)
	if err != nil {
		return nil, fmt.Errorf("failed to hash api key: %w", err)
	}

	quota := input.StorageQuota
	if quota <= 0 {
		quota = 10737418240 // 10GB default
	}

	now := time.Now().UTC()
	u := &user.User{
		ID:           uuid.New(),
		Username:     input.Username,
		Email:        input.Email,
		APIKeyID:     keyID,
		APIKeyHash:   hash,
		StorageQuota: quota,
		StorageUsed:  0,
		Status:       string(user.Active),
		CreatedAt:    &now,
		UpdatedAt:    &now,
	}

	if err := s.repo.Create(ctx, u); err != nil {
		return nil, err
	}

	return &UserDTO{
		ID:           u.ID,
		Username:     u.Username,
		Email:        u.Email,
		APIKey:       fullKey,
		StorageQuota: u.StorageQuota,
		StorageUsed:  u.StorageUsed,
		Status:       u.Status,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}, nil
}

// Get retrieves a user by ID and returns a UserDTO.
func (s *UserService) Get(ctx context.Context, id uuid.UUID) (*UserDTO, error) {
	u, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	return &UserDTO{
		ID:           u.ID,
		Username:     u.Username,
		Email:        u.Email,
		StorageQuota: u.StorageQuota,
		StorageUsed:  u.StorageUsed,
		Status:       u.Status,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}, nil
}

// UpdateStatus changes the status of a user.
func (s *UserService) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	lowerStatus := strings.ToLower(strings.TrimSpace(status))

	userStatus := user.UserStatus(lowerStatus)
	if userStatus != user.Active && userStatus != user.Suspended && userStatus != user.Deleted {
		return fmt.Errorf("invalid status: %s", status)
	}

	return s.repo.UpdateStatus(ctx, id, userStatus)
}

func (s *UserService) validateCreateInput(input CreateUserInput) error {
	if input.Username == "" || input.Email == "" {
		return fmt.Errorf("username and email are required")
	}
	return nil
}
