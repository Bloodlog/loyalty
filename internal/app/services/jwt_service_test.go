package services

import (
	"gophermart/internal/app/dto"
	"gophermart/internal/config"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJwtService(t *testing.T) {
	cfg := &config.Config{
		AuthSecretKey:    "test_secret",
		AuthTokenExpired: time.Hour,
	}

	jwtService := NewJwtService(cfg)

	t.Run("CreateJwt should return a valid token", func(t *testing.T) {
		userID := 123
		tokenString, err := jwtService.CreateJwt(userID)
		require.NoError(t, err)
		require.NotEmpty(t, tokenString)

		claims := &dto.Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.AuthSecretKey), nil
		})

		require.NoError(t, err)
		assert.True(t, token.Valid)
		assert.Equal(t, userID, claims.UserID)
	})

	t.Run("GetUserID should extract correct user ID from a valid token", func(t *testing.T) {
		userID := 456
		tokenString, err := jwtService.CreateJwt(userID)
		require.NoError(t, err)

		extractedUserID, err := jwtService.GetUserID(tokenString)
		require.NoError(t, err)
		assert.Equal(t, userID, extractedUserID)
	})

	t.Run("GetUserID should return error for invalid token", func(t *testing.T) {
		invalidToken := "invalid.token.string"

		userID, err := jwtService.GetUserID(invalidToken)
		assert.Error(t, err)
		assert.Zero(t, userID)
	})
}
