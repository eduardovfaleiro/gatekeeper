package handler

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/eduardovfaleiro/gatekeeper/internal/repository"
	"github.com/eduardovfaleiro/gatekeeper/internal/service"
	authpb "github.com/eduardovfaleiro/gatekeeper/proto"
	"github.com/go-playground/validator/v10"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthHandler struct {
	authpb.UnimplementedAuthServiceServer
	svc service.AuthService
}

func NewAuthHandler(svc service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

var validate = validator.New()

func (h *AuthHandler) Register(ctx context.Context, req *authpb.RegisterRequest) (*authpb.RegisterResponse, error) {
	if err := validate.Var(req.Email, "required,email"); err != nil {
		return nil, fmt.Errorf("invalid email format")
	}
	if err := h.validatePassword(req.Password); err != nil {
		return nil, fmt.Errorf("password length must be between 8 and 32 characteres")
	}

	user, err := h.svc.Register(ctx, req.Email, req.Password)
	if err != nil {
		return nil, err
	}

	return &authpb.RegisterResponse{
		Id:        user.ID.String(),
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (h *AuthHandler) Login(ctx context.Context, req *authpb.LoginRequest) (*authpb.LoginResponse, error) {
	token, err := h.svc.Login(ctx, req.Email, req.Password)

	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}

	return &authpb.LoginResponse{
		AccessToken: token,
	}, nil
}

func (h *AuthHandler) ForgotPassword(ctx context.Context, req *authpb.ForgotPasswordRequest) (*authpb.ForgotPasswordResponse, error) {
	err := h.svc.ForgotPassword(ctx, req.Email)

	if err != nil {
		if !errors.Is(err, repository.ErrNotFound) {
			log.Printf("ERROR: AuthHandler.ForgotPassword failure: %v", err)
			return nil, status.Error(codes.Internal, "internal server error")
		}

		log.Printf("INFO: AuthHandler.ForgotPassword: email not found: %s", req.Email)
	}

	return &authpb.ForgotPasswordResponse{
		Message: "If the email exists in our database, a reset link was sent.",
	}, nil
}

func (h *AuthHandler) ResetPassword(ctx context.Context, req *authpb.ResetPasswordRequest) (*authpb.ResetPasswordResponse, error) {
	if err := validate.Var(req.Token, "required"); err != nil {
		return nil, status.Error(codes.InvalidArgument, "token is required")
	}
	if err := h.validatePassword(req.NewPassword); err != nil {
		return nil, status.Error(codes.InvalidArgument, "password length must be between 8 and 32 characters")
	}

	err := h.svc.ResetPassword(ctx, req.Token, req.NewPassword)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "invalid or expired token")
		}

		log.Printf("ERROR: AuthHandler.ResetPassword failure: %v", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &authpb.ResetPasswordResponse{
		Message: "Password updated successfully",
	}, nil
}

func (h *AuthHandler) validatePassword(password string) error {
	err := validate.Var(password, "required,min=8,max=32")
	if err != nil {
		return status.Error(codes.InvalidArgument, "password length must be between 8 and 32 characters")
	}
	return nil
}
