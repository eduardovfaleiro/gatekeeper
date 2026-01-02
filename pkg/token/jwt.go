package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/eduardovfaleiro/gatekeeper/internal/model"
	"github.com/golang-jwt/jwt/v5"
)

func GenerateToken(userID model.ID, secret string) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID.String(),
		"exp": time.Now().Add(time.Hour * 24).Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func ValidateToken(tokenStr string, secret string) (model.ID, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil || !token.Valid {
		return model.ID{}, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return model.ID{}, errors.New("invalid claims")
	}

	userIDStr, ok := claims["sub"].(string)
	if !ok {
		return model.ID{}, errors.New("subject not found in token")
	}

	return model.ParseID(userIDStr)
}
