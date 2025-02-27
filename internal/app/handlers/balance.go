package handlers

import (
	"encoding/json"
	"errors"
	"gophermart/internal/app/apperrors"
	"gophermart/internal/app/dto"
	"gophermart/internal/app/services"
	"gophermart/internal/app/utils"
	"net/http"

	"go.uber.org/zap"
)

type BalanceHandler struct {
	BalanceService services.BalanceService
	Logger         *zap.SugaredLogger
}

func NewBalanceHandler(balanceService services.BalanceService, logger *zap.SugaredLogger) *BalanceHandler {
	handlerLogger := logger.With("component:NewBalanceHandler", "BalanceHandler")
	return &BalanceHandler{
		BalanceService: balanceService,
		Logger:         handlerLogger,
	}
}

func (b *BalanceHandler) GetUserBalance() http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		ctx := request.Context()
		userID, err := utils.GetUserID(ctx)
		if err != nil {
			response.WriteHeader(http.StatusUnauthorized)
			return
		}
		balance, err := b.BalanceService.GetBalance(ctx, userID)
		if err != nil {
			response.WriteHeader(http.StatusInternalServerError)
			return
		}
		response.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(response).Encode(balance)
		if err != nil {
			b.Logger.Infoln("error Encode balance", err)
			response.WriteHeader(http.StatusInternalServerError)
			return
		}
		response.WriteHeader(http.StatusOK)
	}
}

func (b *BalanceHandler) StoreBalanceWithdraw() http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		ctx := request.Context()
		userID, err := utils.GetUserID(ctx)
		if err != nil {
			response.WriteHeader(http.StatusUnauthorized)
			return
		}
		var req dto.WithdrawBody
		if err = json.NewDecoder(request.Body).Decode(&req); err != nil {
			response.WriteHeader(http.StatusBadRequest)
			return
		}
		if err = b.BalanceService.Withdraw(ctx, userID, req); err != nil {
			if errors.Is(err, apperrors.ErrDuplicateOrderID) {
				// номер заказа уже был загружен этим пользователем;
				response.WriteHeader(http.StatusOK)
			}
			if errors.Is(err, apperrors.ErrBalanceNotEnought) {
				// на счету недостаточно средств;
				response.WriteHeader(http.StatusPaymentRequired)
			}
			b.Logger.Infoln("error save withdraw", err)
			response.WriteHeader(http.StatusInternalServerError)
			return
		}
		response.WriteHeader(http.StatusOK)
	}
}

func (b *BalanceHandler) GetWithdrawals() http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		ctx := request.Context()
		userID, err := utils.GetUserID(ctx)
		if err != nil {
			response.WriteHeader(http.StatusUnauthorized)
			return
		}
		withdrawals, err := b.BalanceService.GetWithdrawals(ctx, userID)
		if err != nil {
			b.Logger.Infoln("error GetWithdrawals", err)
			response.WriteHeader(http.StatusInternalServerError)
			return
		}
		response.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(response).Encode(withdrawals)
		if err != nil {
			b.Logger.Infoln("error Encode withdrawals", err)
			response.WriteHeader(http.StatusInternalServerError)
			return
		}
		response.WriteHeader(http.StatusOK)
	}
}
