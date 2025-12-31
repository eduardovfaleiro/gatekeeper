package service

import (
	"context"
	"fmt"

	"github.com/eduardovfaleiro/gatekeeper/internal/model"
	"github.com/eduardovfaleiro/gatekeeper/internal/repository"
	"github.com/eduardovfaleiro/gatekeeper/pkg/hash"
	passwordValidator "github.com/wagslane/go-password-validator"
)

type AuthService interface {
	Register(ctx context.Context, email, password string) (*model.User, error)
}

type authService struct {
	repo repository.UserRepository
}

func NewAuthService(repo repository.UserRepository) AuthService {
	return &authService{repo: repo}
}

func (s *authService) Register(ctx context.Context, email, password string) (*model.User, error) {
	const minEntropyBits = 60
	err := passwordValidator.Validate(password, minEntropyBits)
	if err != nil {
		return nil, fmt.Errorf("password too weak: %v", err)
	}

	hashedPassword, err := hash.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Email:        email,
		PasswordHash: hashedPassword,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}
