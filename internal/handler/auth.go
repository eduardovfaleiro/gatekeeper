package handler

import (
	"context"
	"fmt"
	"time"

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
	if err := validate.Var(req.Password, "required,min=8,max=32"); err != nil {
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
