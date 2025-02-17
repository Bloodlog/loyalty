package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"loyalty/internal/app/apperrors"
	"loyalty/internal/app/dto"
	"loyalty/internal/app/entities"
	"loyalty/internal/app/services"
	"net/http"
	"strconv"
	"strings"
)

type OrderHandler struct {
	OrderService services.OrderService
}

func NewOrderHandler(orderService services.OrderService) *OrderHandler {
	return &OrderHandler{OrderService: orderService}
}

func (h *OrderHandler) GetUserOrders() http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		ctx := request.Context()
		userID := ctx.Value("userID").(int)
		orders, err := h.OrderService.GetOrdersByUserID(ctx, userID)
		if err != nil {
			fmt.Println(err)
			response.WriteHeader(http.StatusNoContent)
			return
		}
		if orders == nil {
			orders = []dto.OrdersResponseBody{}
		}
		response.WriteHeader(http.StatusOK)
		err = json.NewEncoder(response).Encode(orders)
		if err != nil {
			response.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func (h *OrderHandler) StoreOrders() http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		ctx := request.Context()
		userID := ctx.Value("userID").(int)
		body, err := io.ReadAll(request.Body)
		if err != nil {
			response.WriteHeader(http.StatusBadRequest)
			return
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				return
			}
		}(request.Body)

		numberStr := strings.TrimSpace(string(body))

		number, err := strconv.ParseInt(numberStr, 10, 64)
		if err != nil {
			response.WriteHeader(http.StatusBadRequest)
			return
		}

		if !luhnCheck(number) {
			response.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		var req dto.OrderBody

		req.OrderNumber = number
		req.UserId = int64(userID)
		req.StatusId = entities.StatusNew

		err = h.OrderService.SaveOrder(ctx, req)
		if err != nil {
			if errors.Is(err, apperrors.ErrDuplicateOrderID) {
				// номер заказа уже был загружен этим пользователем;
				response.WriteHeader(http.StatusOK)
			}
			//номер заказа уже был загружен другим пользователем;
			//response.WriteHeader(http.StatusConflict)
			response.WriteHeader(http.StatusInternalServerError)
		}

		//новый номер заказа принят в обработку;
		response.WriteHeader(http.StatusAccepted)
	}
}

func luhnCheck(number int64) bool {
	numStr := strconv.FormatInt(number, 10)
	length := len(numStr)

	return length >= 8 && length <= 19
}
