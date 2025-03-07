package handlers

import (
	"encoding/json"
	"errors"
	"gophermart/internal/app/apperrors"
	"gophermart/internal/app/dto"
	"gophermart/internal/app/entities"
	"gophermart/internal/app/services"
	"net/http"

	"go.uber.org/zap"
)

type UserHandler struct {
	UserService services.UserService
	JwtService  services.JwtService
	Logger      *zap.SugaredLogger
}

func NewUserHandler(
	userService services.UserService,
	jwtService services.JwtService,
	logger *zap.SugaredLogger,
) *UserHandler {
	handlerLogger := logger.With("component:NewUserHandler", "UserHandler")
	return &UserHandler{
		UserService: userService,
		JwtService:  jwtService,
		Logger:      handlerLogger,
	}
}

func (u *UserHandler) LoginUser() http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		ctx := request.Context()
		var req dto.LoginRequestBody
		if err := json.NewDecoder(request.Body).Decode(&req); err != nil {
			response.WriteHeader(http.StatusBadRequest)
			return
		}

		var user entities.User
		var err error
		user, err = u.UserService.Login(ctx, req)
		if err != nil {
			response.WriteHeader(http.StatusUnauthorized)
			return
		}

		tokenString, err := u.JwtService.CreateJwt(user.ID)
		if err != nil {
			u.Logger.Infoln("error CreateJwt", err)
			response.WriteHeader(http.StatusInternalServerError)
			return
		}
		cookie := &http.Cookie{
			Name:     "Authorization",
			Value:    tokenString,
			Path:     "/",
			HttpOnly: false,
			Secure:   false,
		}
		http.SetCookie(response, cookie)
		response.Header().Set("Authorization", tokenString)
		err = json.NewEncoder(response).Encode(map[string]string{"token": tokenString})
		if err != nil {
			u.Logger.Infoln("error Encode token", err)
			response.WriteHeader(http.StatusInternalServerError)
			return
		}
		response.WriteHeader(http.StatusOK)
	}
}

func (u *UserHandler) RegisterUser() http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		ctx := request.Context()
		var req dto.RegisterRequestBody
		if err := json.NewDecoder(request.Body).Decode(&req); err != nil {
			response.WriteHeader(http.StatusBadRequest)
			return
		}

		var user entities.User
		var err error
		user, err = u.UserService.Register(ctx, req)
		if err != nil {
			if errors.Is(err, apperrors.ErrDuplicateLogin) {
				// логин уже занят;
				response.WriteHeader(http.StatusConflict)
			}
			u.Logger.Infoln("error Register", err)
			response.WriteHeader(http.StatusInternalServerError)
			return
		}

		tokenString, err := u.JwtService.CreateJwt(user.ID)
		if err != nil {
			u.Logger.Infoln("error CreateJwt", err)
			response.WriteHeader(http.StatusInternalServerError)
			return
		}

		cookie := &http.Cookie{
			Name:     "Authorization",
			Value:    tokenString,
			Path:     "/",
			HttpOnly: false,
			Secure:   false,
		}
		http.SetCookie(response, cookie)
		response.Header().Set("Authorization", tokenString)

		err = json.NewEncoder(response).Encode(map[string]string{"token": tokenString})
		if err != nil {
			u.Logger.Infoln("error Encode token", err)
			response.WriteHeader(http.StatusInternalServerError)
			return
		}
		response.WriteHeader(http.StatusOK)
	}
}
