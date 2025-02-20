package handlers

import (
	"encoding/json"
	"errors"
	"loyalty/internal/app/apperrors"
	"loyalty/internal/app/dto"
	"loyalty/internal/app/entities"
	"loyalty/internal/app/services"
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

func (h *UserHandler) LoginUser() http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		ctx := request.Context()
		var req dto.LoginRequestBody
		if err := json.NewDecoder(request.Body).Decode(&req); err != nil {
			response.WriteHeader(http.StatusBadRequest)
			return
		}

		var user entities.User
		var err error
		user, err = h.UserService.Login(ctx, req)
		if err != nil {
			response.WriteHeader(http.StatusUnauthorized)
			return
		}

		tokenString, err := h.JwtService.CreateJwt(user.ID)
		if err != nil {
			h.Logger.Infoln("error CreateJwt", err)
			response.WriteHeader(http.StatusInternalServerError)
			return
		}
		err = json.NewEncoder(response).Encode(map[string]string{"token": tokenString})
		if err != nil {
			h.Logger.Infoln("error Encode token", err)
			response.WriteHeader(http.StatusInternalServerError)
			return
		}

		response.Header().Set("Authorization", tokenString)
		response.WriteHeader(http.StatusOK)
	}
}

func (h *UserHandler) RegisterUser() http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		ctx := request.Context()
		var req dto.RegisterRequestBody
		if err := json.NewDecoder(request.Body).Decode(&req); err != nil {
			response.WriteHeader(http.StatusBadRequest)
			return
		}

		var user entities.User
		var err error
		user, err = h.UserService.Register(ctx, req)
		if err != nil {
			if errors.Is(err, apperrors.ErrDuplicateLogin) {
				// логин уже занят;
				response.WriteHeader(http.StatusConflict)
			}
			h.Logger.Infoln("error Register", err)
			response.WriteHeader(http.StatusInternalServerError)
			return
		}

		tokenString, err := h.JwtService.CreateJwt(user.ID)
		if err != nil {
			h.Logger.Infoln("error CreateJwt", err)
			response.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = json.NewEncoder(response).Encode(map[string]string{"token": tokenString})
		if err != nil {
			h.Logger.Infoln("error Encode token", err)
			response.WriteHeader(http.StatusInternalServerError)
			return
		}

		response.Header().Set("Authorization", tokenString)
		response.WriteHeader(http.StatusOK)
	}
}
