package routers

import (
	"loyalty/internal/app/handlers"
	"loyalty/internal/app/repositories"
	"loyalty/internal/app/services"
	"loyalty/internal/config"
	"loyalty/internal/middlewares"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func ConfigureServerHandler(db *pgxpool.Pool, cfg *config.Config) http.Handler {
	router := chi.NewRouter()

	router.Use(middleware.Logger)

	registerAPIRouter(router, db, cfg)

	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	return router
}

func registerAPIRouter(r *chi.Mux, db *pgxpool.Pool, cfg *config.Config) {
	userRepo := repositories.NewUserRepository(db)
	orderRepo := repositories.NewOrderRepository(db)
	withdrawRepo := repositories.NewWithdrawRepository(db)

	userService := services.NewUserService(userRepo)
	orderService := services.NewOrderService(orderRepo)
	balanceService := services.NewBalanceService(orderRepo, withdrawRepo)
	jwtService := services.NewJwtService(cfg)

	userHandler := handlers.NewUserHandler(userService, jwtService)
	orderHandler := handlers.NewOrderHandler(orderService)
	balanceHandler := handlers.NewBalanceHandler(balanceService)

	r.Route("/api/user", func(r chi.Router) {
		r.Post("/register", userHandler.RegisterUser())
		r.Post("/login", userHandler.LoginUser())
		r.Group(func(r chi.Router) {
			r.Use(middlewares.Auth(jwtService, userRepo))
			r.Post("/orders", orderHandler.StoreOrders())
			r.Get("/orders", orderHandler.GetUserOrders())

			r.Get("/balance", balanceHandler.GetUserBalance())
			r.Post("/balance/withdraw", balanceHandler.StoreBalanceWithdraw())
			r.Get("/withdrawals", balanceHandler.GetWithdrawals())
		})
	})
}
