package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/eduardovfaleiro/gatekeeper/internal/model"
	"github.com/eduardovfaleiro/gatekeeper/internal/repository"
	"github.com/eduardovfaleiro/gatekeeper/pkg/hash"
	"github.com/eduardovfaleiro/gatekeeper/pkg/token"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	passwordValidator "github.com/wagslane/go-password-validator"
)

type AuthService interface {
	Register(ctx context.Context, email, password string) (*model.User, error)
	Login(ctx context.Context, email, password string) (string, error)
	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token, new_password string) error
}

type authService struct {
	repo         repository.UserRepository
	redis        *redis.Client
	emailService EmailService
}

func NewAuthService(repo repository.UserRepository, redis *redis.Client, emailService EmailService) AuthService {
	return &authService{repo: repo, redis: redis, emailService: emailService}
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
		ID:           model.NewID(),
		Email:        email,
		PasswordHash: hashedPassword,
		CreatedAt:    model.NewTimestamp(),
	}

	err = s.repo.Create(ctx, user)

	if err != nil {
		return nil, fmt.Errorf("could not create user: %w", err)
	}

	return user, nil
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

func (s *authService) ForgotPassword(ctx context.Context, email string) error {
	user, err := s.repo.GetByEmail(ctx, email)

	if err != nil {
		return err
	}

	resetToken := uuid.New().String()

	key := fmt.Sprintf("password_reset:%s", resetToken)
	err = s.redis.Set(ctx, key, user.ID.String(), 15*time.Minute).Err()

	if err != nil {
		return fmt.Errorf("authService.ForgotPassword (redis set): %w", err)
	}

	go func() {
		err := s.emailService.SendResetLink(user.Email, resetToken)
		if err != nil {
			log.Printf("ERROR: authService.ForgotPassword background email: %v", err)
		}
	}()

	return nil
}

func (s *authService) ResetPassword(ctx context.Context, token, newPassword string) error {
	key := fmt.Sprintf("password_reset:%s", token)

	userIDStr, err := s.redis.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return repository.ErrNotFound
		}
		return fmt.Errorf("authService.ResetPassword (redis get): %w", err)
	}

	userID, err := model.ParseID(userIDStr)
	if err != nil {
		return fmt.Errorf("authService.ResetPassword (parseID): %w", err)
	}

	hashedPassword, err := hash.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("authService.ResetPassword (hash): %w", err)
	}

	err = s.repo.UpdatePassword(ctx, userID, hashedPassword)
	if err != nil {
		return fmt.Errorf("authService.ResetPassword (repo): %w", err)
	}

	err = s.redis.Del(ctx, key).Err()
	if err != nil {
		log.Printf("WARN: failed to delete reset token from redis: %v", err)
	}

	return nil
}
