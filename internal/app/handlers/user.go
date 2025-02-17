package handlers

import (
	"encoding/json"
	"errors"
	"loyalty/internal/app/apperrors"
	"loyalty/internal/app/dto"
	"loyalty/internal/app/entities"
	"loyalty/internal/app/services"
	"net/http"
)

type UserHandler struct {
	UserService services.UserService
	JwtService  services.JwtService
}

func NewUserHandler(
	userService services.UserService,
	jwtService services.JwtService,
) *UserHandler {
	return &UserHandler{
		UserService: userService,
		JwtService:  jwtService,
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
			response.WriteHeader(http.StatusInternalServerError)
			return
		}
		response.WriteHeader(http.StatusOK)
		json.NewEncoder(response).Encode(map[string]string{"token": tokenString})
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
			response.WriteHeader(http.StatusInternalServerError)
			return
		}

		tokenString, err := h.JwtService.CreateJwt(user.ID)
		if err != nil {
			response.WriteHeader(http.StatusInternalServerError)
			return
		}

		response.WriteHeader(http.StatusOK)
		json.NewEncoder(response).Encode(map[string]string{"token": tokenString})
	}
}
