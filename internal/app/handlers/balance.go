package handlers

import (
	"encoding/json"
	"errors"
	"loyalty/internal/app/apperrors"
	"loyalty/internal/app/dto"
	"loyalty/internal/app/services"
	"net/http"
)

type BalanceHandler struct {
	BalanceService services.BalanceService
}

func NewBalanceHandler(balanceService services.BalanceService) *BalanceHandler {
	return &BalanceHandler{BalanceService: balanceService}
}

func (h *BalanceHandler) GetUserBalance() http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		ctx := request.Context()
		userID := ctx.Value("userID").(int)
		balance, err := h.BalanceService.GetBalance(ctx, userID)
		if err != nil {
			response.WriteHeader(http.StatusInternalServerError)
			return
		}
		response.WriteHeader(http.StatusOK)
		json.NewEncoder(response).Encode(balance)
	}
}

func (h *BalanceHandler) StoreBalanceWithdraw() http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		var req dto.WithdrawBody
		if err := json.NewDecoder(request.Body).Decode(&req); err != nil {
			response.WriteHeader(http.StatusBadRequest)
			return
		}
		ctx := request.Context()
		userID := ctx.Value("userID").(int)

		if err := h.BalanceService.Withdraw(ctx, userID, req); err != nil {
			if errors.Is(err, apperrors.ErrDuplicateOrderID) {
				// номер заказа уже был загружен этим пользователем;
				response.WriteHeader(http.StatusOK)
			}
			if errors.Is(err, apperrors.ErrBalanceNotEnought) {
				// на счету недостаточно средств;
				response.WriteHeader(http.StatusPaymentRequired)
			}
			response.WriteHeader(http.StatusInternalServerError)
			return
		}

		response.WriteHeader(http.StatusOK)
	}
}

func (h *BalanceHandler) GetWithdrawals() http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		ctx := request.Context()
		userID := ctx.Value("userID").(int)

		withdrawals, err := h.BalanceService.GetWithdrawals(ctx, userID)
		if err != nil {
			response.WriteHeader(http.StatusInternalServerError)
			return
		}

		response.WriteHeader(http.StatusOK)
		json.NewEncoder(response).Encode(withdrawals)
	}
}
