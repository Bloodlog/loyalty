package handlers

import (
	"context"
	"loyalty/internal/app/entities"
	"loyalty/internal/app/services"
	"loyalty/internal/config"
	"sync"
	"time"

	"go.uber.org/zap"
)

type SendOrderHandler struct {
	SendOrderService services.SendOrdersService
	Cfg              *config.Config
	sendQueue        chan entities.Order
	Logger           *zap.SugaredLogger
}

func NewSendOrderHandler(
	sendOrderService services.SendOrdersService,
	cfg *config.Config,
	logger *zap.SugaredLogger,
) *SendOrderHandler {
	return &SendOrderHandler{
		SendOrderService: sendOrderService,
		Cfg:              cfg,
		Logger:           logger,
	}
}

func (h *SendOrderHandler) SendUserOrders() error {
	pollTicker := time.NewTicker(h.Cfg.PollInterval)
	defer pollTicker.Stop()

	h.sendQueue = make(chan entities.Order, h.Cfg.RateLimit)

	var wg sync.WaitGroup
	for i := 0; i < h.Cfg.RateLimit; i++ {
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

			for _, order := range newOrders {
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
		err := h.SendOrderService.SendOrder(ctx, order)
		if err != nil {
			h.Logger.Infof("Failed to send order %d: %v\n", order.OrderId, err)
		}
	}
}
