package service

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/eduardovfaleiro/gatekeeper/internal/model"
	"github.com/eduardovfaleiro/gatekeeper/internal/repository"
	"github.com/eduardovfaleiro/gatekeeper/pkg/hash"
	"github.com/eduardovfaleiro/gatekeeper/pkg/token"
	passwordValidator "github.com/wagslane/go-password-validator"
)

type AuthService interface {
	Register(ctx context.Context, email, password string) (*model.User, error)
	Login(ctx context.Context, email, password string) (string, error)
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

	u, err := s.repo.Create(ctx, &model.User{
		Email:        email,
		PasswordHash: hashedPassword,
	})

	if err != nil {
		return nil, fmt.Errorf("could not create user: %w", err)
	}

	return u, nil
}

func (s *authService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.repo.GetByEmail(ctx, email)

	if err != nil {
		return "", err
	}

	if !hash.CheckPasswordHash(password, user.PasswordHash) {
		return "", errors.New("invalid credentials")
	}

	secret := os.Getenv("JWT_SECRET")
	token, err := token.GenerateToken(user.ID, secret)

	if err != nil {
		return "", err
	}

	return token, nil
}
