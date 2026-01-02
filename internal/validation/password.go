package validation

import (
	"github.com/go-playground/validator/v10"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var validate = validator.New()

func ValidatePassword(password string) error {
	if err := validate.Var(password, "required,min=8,max=32"); err != nil {
		return status.Error(codes.InvalidArgument, "password length must be between 8 and 32 characters")
	}
	return nil
}
