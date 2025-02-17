package services

import (
	"context"
	"loyalty/internal/app/dto"
	"loyalty/internal/app/entities"
	"loyalty/internal/app/repositories"

	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	Register(ctx context.Context, req dto.RegisterRequestBody) (entities.User, error)
	Login(ctx context.Context, req dto.LoginRequestBody) (entities.User, error)
}

type userService struct {
	UserRepository repositories.UserRepositoryInterface
}

func NewUserService(
	userRepository repositories.UserRepositoryInterface,
) UserService {
	return &userService{
		UserRepository: userRepository,
	}
}

func (u *userService) Register(ctx context.Context, req dto.RegisterRequestBody) (entities.User, error) {
	var user entities.User
	user.Login = req.Login
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return user, err
	}
	user.Password = string(hashedPassword)
	newUser, err := u.UserRepository.Store(ctx, user)
	if err != nil {
		return user, err
	}

	return newUser, nil
}

func (u *userService) Login(ctx context.Context, req dto.LoginRequestBody) (entities.User, error) {
	user, err := u.UserRepository.GetByLogin(ctx, req.Login)
	if err != nil {
		return user, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return user, err
	}

	return user, nil
}
