package services

import (
	"fmt"
	"loyalty/internal/app/dto"
	"loyalty/internal/config"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type JwtService interface {
	CreateJwt(userId int) (string, error)
	GetUserID(tokenString string) (int, error)
}

type jwtService struct {
	Cfg *config.Config
}

func NewJwtService(cfg *config.Config) JwtService {
	return &jwtService{
		Cfg: cfg,
	}
}

func (o *jwtService) CreateJwt(userId int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, dto.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(o.Cfg.AuthTokenExpired)),
		},
		UserID: userId,
	})

	tokenString, err := token.SignedString([]byte(o.Cfg.AuthSecretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (o *jwtService) GetUserID(tokenString string) (int, error) {
	claims := &dto.Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(o.Cfg.AuthSecretKey), nil
		})
	if err != nil {
		return 0, err
	}

	if !token.Valid {
		return 0, err
	}

	return claims.UserID, nil
}
