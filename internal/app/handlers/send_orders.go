package handlers

import (
	"context"
	"errors"
	"loyalty/internal/app/entities"
	"loyalty/internal/app/services"
	"loyalty/internal/app/services/accrual"
	"loyalty/internal/config"
	"sync"
	"time"

	"go.uber.org/zap"
)

type SendOrderHandler struct {
	SendOrderService services.AccrualService
	Cfg              *config.Config
	sendQueue        chan *entities.Order
	Logger           *zap.SugaredLogger
	mu               sync.Mutex
}

func NewSendOrderHandler(
	sendOrderService services.AccrualService,
	cfg *config.Config,
	queue chan *entities.Order,
	logger *zap.SugaredLogger,
) *SendOrderHandler {
	handlerLogger := logger.With("component:NewSendOrderHandler", "SendOrderHandler")
	return &SendOrderHandler{
		SendOrderService: sendOrderService,
		Cfg:              cfg,
		Logger:           handlerLogger,
		sendQueue:        queue,
	}
}

func (h *SendOrderHandler) SendUserOrders() error {
	pollTicker := time.NewTicker(h.Cfg.PollInterval)
	defer pollTicker.Stop()

	var wg sync.WaitGroup
	for range make([]struct{}, h.Cfg.RateLimit) {
		wg.Add(1)
		go h.worker(&wg)
	}

	go func() {
		ctx := context.Background()
		for range pollTicker.C {
			newOrders, err := h.SendOrderService.GetFreshOrders(ctx, h.Cfg.AgentOrderLimit)
			if err != nil {
				h.Logger.Infof("Error getting new orders: %v\n", err)
				continue
			}

			for i := range newOrders {
				order := &newOrders[i]
				h.sendQueue <- order
			}
		}
	}()

	wg.Wait()
	close(h.sendQueue)

	return nil
}

func (h *SendOrderHandler) worker(wg *sync.WaitGroup) {
	defer wg.Done()

	for order := range h.sendQueue {
		ctx := context.Background()
		h.mu.Lock()
		err := h.SendOrderService.SendOrder(ctx, order)
		h.mu.Unlock()

		if err != nil {
			var tooManyReqErr *accrual.ErrTooManyRequestsWithRetry
			if errors.As(err, &tooManyReqErr) {
				h.Logger.Infof("Слишком много запросов, пауза %d секунд\n", tooManyReqErr.RetryAfter)

				h.mu.Lock()
				time.Sleep(time.Duration(tooManyReqErr.RetryAfter) * time.Second)
				h.mu.Unlock()
			} else {
				h.Logger.Infof("Failed to send order %d: %v\n", order.OrderID, err)
			}
		}
	}
}
