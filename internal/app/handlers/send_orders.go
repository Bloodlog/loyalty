package handlers

import (
	"context"
	"errors"
	"gophermart/internal/app/entities"
	"gophermart/internal/app/services"
	"gophermart/internal/app/services/accrual"
	"gophermart/internal/config"
	"sync"
	"time"

	"go.uber.org/zap"
)

type SendOrderHandler struct {
	SendOrderService services.AccrualService
	Cfg              *config.Config
	sendQueue        chan *entities.Job
	Logger           *zap.SugaredLogger
	mu               *sync.RWMutex
}

func NewSendOrderHandler(
	sendOrderService services.AccrualService,
	cfg *config.Config,
	logger *zap.SugaredLogger,
) *SendOrderHandler {
	handlerLogger := logger.With("component:NewSendOrderHandler", "SendOrderHandler")
	sendQueue := make(chan *entities.Job, cfg.RateLimit)
	return &SendOrderHandler{
		SendOrderService: sendOrderService,
		Cfg:              cfg,
		Logger:           handlerLogger,
		sendQueue:        sendQueue,
		mu:               &sync.RWMutex{},
	}
}

func (h *SendOrderHandler) SendUserOrders() error {
	ctx := context.Background()
	timer := time.NewTimer(h.Cfg.PollInterval)
	defer timer.Stop()

	var wg sync.WaitGroup
	for range make([]struct{}, h.Cfg.RateLimit) {
		wg.Add(1)
		go h.worker(ctx, &wg, timer)
	}

	select {
	case <-ctx.Done():
		h.Logger.Info("Shutting down gracefully...")
		return nil
	case <-timer.C:
		jobs, err := h.SendOrderService.GetPendingJobs(ctx, h.Cfg.AgentOrderLimit)
		if err != nil {
			h.Logger.Errorf("Failed to get jobs: %v", err)
			timer.Reset(h.Cfg.PollInterval)
			break
		}
		if len(jobs) == 0 {
			timer.Reset(h.Cfg.PollInterval)
			break
		}
		for _, job := range jobs {
			select {
			case h.sendQueue <- &job:
			case <-ctx.Done():
				return nil
			}
		}
	}

	wg.Wait()
	close(h.sendQueue)

	return nil
}

func (h *SendOrderHandler) worker(ctx context.Context, wg *sync.WaitGroup, timer *time.Timer) {
	defer wg.Done()

	for job := range h.sendQueue {
		h.mu.Lock()
		err := h.SendOrderService.SendOrder(ctx, job)
		h.mu.Unlock()

		if err != nil {
			var tooManyReqErr *accrual.TooManyRequestsWithRetryError
			if errors.As(err, &tooManyReqErr) {
				h.Logger.Infof("Слишком много запросов, пауза %d секунд\n", tooManyReqErr.RetryAfter)
				after := time.Duration(tooManyReqErr.RetryAfter) * time.Second
				timer.Reset(after)
				<-timer.C
			} else {
				h.Logger.Infof("Failed job with order id %d: %v\n", job.OrderID, err)
			}
		}
	}
}
