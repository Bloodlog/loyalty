package services

import (
	"context"
	"gophermart/internal/app/dto"
	"gophermart/internal/app/entities"
	"gophermart/internal/app/mocks"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestLogin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()

	m := mocks.NewMockUserRepositoryInterface(ctrl)

	login := "user1"
	password := "password"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	m.EXPECT().GetByLogin(ctx, login).Return(entities.User{
		Login:    login,
		Password: string(hashedPassword),
	}, nil)

	userService := NewUserService(m)

	req := dto.LoginRequestBody{
		Login:    login,
		Password: password,
	}

	user, err := userService.Login(ctx, req)
	require.NoError(t, err)
	require.Equal(t, user.Password, string(hashedPassword))
	require.Equal(t, user.Login, login)
}
