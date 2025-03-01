package services

import (
	"context"
	"database/sql"
	"fmt"
	"gophermart/internal/app/entities"
	"gophermart/internal/app/services/accrual"
	"math"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"

	"gophermart/internal/app/repositories"
	"gophermart/internal/config"
)

type AccrualService interface {
	SendOrder(ctx context.Context, job *entities.Job) error
	GetFreshOrders(ctx context.Context, limit int) ([]entities.Order, error)
	GetPendingJobs(ctx context.Context, limit int) ([]entities.Job, error)
}

type accrualService struct {
	Pool            *pgxpool.Pool
	JobRepository   repositories.JobRepositoryInterface
	OrderRepository repositories.OrderRepositoryInterface
	Client          *resty.Client
	Cfg             *config.Config
	Logger          *zap.SugaredLogger
	roundingFactor  float64
}

func NewAccrualService(
	db *pgxpool.Pool,
	jobRepository repositories.JobRepositoryInterface,
	orderRepository repositories.OrderRepositoryInterface,
	client *resty.Client,
	cfg *config.Config,
	logger *zap.SugaredLogger,
) AccrualService {
	const roundingFactor = 100
	return &accrualService{
		Pool:            db,
		JobRepository:   jobRepository,
		OrderRepository: orderRepository,
		Client:          client,
		Cfg:             cfg,
		Logger:          logger,
		roundingFactor:  roundingFactor,
	}
}

func (a *accrualService) GetFreshOrders(ctx context.Context, limit int) ([]entities.Order, error) {
	orders, err := a.OrderRepository.GetFreshOrders(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to GetFreshOrders: %w", err)
	}
	return orders, nil
}

func (a *accrualService) SendOrder(ctx context.Context, job *entities.Job) error {
	tx, err := a.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func(tx pgx.Tx, ctx context.Context) {
		_ = tx.Rollback(ctx)
	}(tx, ctx)

	order, err := a.OrderRepository.GetByID(ctx, tx, job.OrderID)
	if err != nil {
		return fmt.Errorf("failed to GetById to accrual: %w", err)
	}

	orderResponse, err := accrual.SendOrder(a.Client, order.OrderID)
	if err != nil {
		a.Logger.Infoln(err)
		err = a.JobRepository.UpdateJobPoolAt(ctx, tx, job.ID)
		return fmt.Errorf("failed to SendOrder to accrual: %w", err)
	}

	statusID, _ := entities.GetStatusIDByName(orderResponse.Status)
	if orderResponse.Accrual != nil {
		roundedAmount := math.Round(*orderResponse.Accrual*a.roundingFactor) / a.roundingFactor
		order.Accrual = sql.NullFloat64{Float64: roundedAmount, Valid: true}
	} else {
		order.Accrual = sql.NullFloat64{Valid: false}
	}
	order.StatusID = int16(statusID)
	order.UpdatedAt = time.Now()
	err = a.OrderRepository.UpdateOrder(ctx, tx, order)
	if err != nil {
		a.Logger.Infoln(err)
		return fmt.Errorf("failed to UpdateOrder with status: %w", err)
	}

	err = a.JobRepository.DeleteJobByID(ctx, tx, job.ID)
	if err != nil {
		a.Logger.Infoln(err)
		return fmt.Errorf("failed to DeleteJobByID: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (a *accrualService) GetPendingJobs(ctx context.Context, limit int) ([]entities.Job, error) {
	jobs, err := a.JobRepository.GetPendingJobs(ctx, nil, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to GetPendingJobs: %w", err)
	}
	return jobs, nil
}
