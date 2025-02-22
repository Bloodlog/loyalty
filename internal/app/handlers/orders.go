package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"loyalty/internal/app/apperrors"
	"loyalty/internal/app/dto"
	"loyalty/internal/app/entities"
	"loyalty/internal/app/services"
	"loyalty/internal/app/utils"
	"net/http"
	"strconv"
	"strings"

	"go.uber.org/zap"
)

type OrderHandler struct {
	OrderService services.OrderService
	Logger       *zap.SugaredLogger
}

func NewOrderHandler(orderService services.OrderService, logger *zap.SugaredLogger) *OrderHandler {
	handlerLogger := logger.With("component:NewOrderHandler", "OrderHandler")
	return &OrderHandler{
		OrderService: orderService,
		Logger:       handlerLogger,
	}
}

func (h *OrderHandler) GetUserOrders() http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		ctx := request.Context()
		userID, err := utils.GetUserID(ctx)
		if err != nil {
			response.WriteHeader(http.StatusUnauthorized)
			return
		}
		orders, err := h.OrderService.GetOrdersByUserID(ctx, userID)
		if err != nil {
			response.WriteHeader(http.StatusNoContent)
			return
		}
		if orders == nil {
			orders = []dto.OrdersResponseBody{}
		}
		response.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(response).Encode(orders)
		if err != nil {
			h.Logger.Infoln("error Encode orders", err)
			response.WriteHeader(http.StatusInternalServerError)
			return
		}
		response.WriteHeader(http.StatusOK)
	}
}

func (h *OrderHandler) StoreOrders() http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		ctx := request.Context()
		userID, err := utils.GetUserID(ctx)
		if err != nil {
			response.WriteHeader(http.StatusUnauthorized)
			return
		}
		body, err := io.ReadAll(request.Body)
		if err != nil {
			response.WriteHeader(http.StatusBadRequest)
			return
		}
		defer func(Body io.ReadCloser) {
			err = Body.Close()
			if err != nil {
				h.Logger.Infoln("error close body", err)
				response.WriteHeader(http.StatusInternalServerError)
				return
			}
		}(request.Body)

		numberStr := strings.TrimSpace(string(body))
		number, err := strconv.ParseInt(numberStr, 10, 64)
		if err != nil {
			response.WriteHeader(http.StatusBadRequest)
			return
		}

		if !utils.LuhnCheck(number) {
			response.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		var req dto.OrderBody
		req.OrderNumber = number
		req.UserID = int64(userID)
		req.StatusID = entities.StatusNew
		err = h.OrderService.SaveOrder(ctx, req)
		if err != nil {
			if errors.Is(err, apperrors.ErrDuplicateOrderID) {
				response.WriteHeader(http.StatusOK)
				return
			}
			if errors.Is(err, apperrors.ErrDuplicateOrderIDAnotherUserID) {
				response.WriteHeader(http.StatusConflict)
				return
			}
			h.Logger.Infoln("error SaveOrder", err)
			response.WriteHeader(http.StatusInternalServerError)
			return
		}
		response.WriteHeader(http.StatusAccepted)
	}
}
