package main

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/mkarulina/loyalty-system-service.git/config"
	"github.com/mkarulina/loyalty-system-service.git/internal/accrual"
	"github.com/mkarulina/loyalty-system-service.git/internal/authentication"
	"github.com/mkarulina/loyalty-system-service.git/internal/handlers"
	middleware2 "github.com/mkarulina/loyalty-system-service.git/internal/middleware"
	"github.com/mkarulina/loyalty-system-service.git/internal/storage"
	"github.com/mkarulina/loyalty-system-service.git/sql"
	"github.com/spf13/viper"
	"log"
	"net/http"
)

func main() {
	_, err := config.LoadConfig("config")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	sql.RunMigration()
	accrual.StartCron()

	h := handlers.NewHandler(
		storage.NewOrderStorage(),
		storage.NewHistoryStorage(),
		authentication.New(),
	)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/api/", func(r chi.Router) {

		r.Route("/user/register", func(r chi.Router) {
			r.Use(middleware2.TokenHandle)
			r.Post("/", h.RegisterHandler)
		})

		r.Route("/user/login", func(r chi.Router) {
			r.Use(middleware2.TokenHandle)
			r.Post("/", h.LoginHandler)
		})

		r.Route("/user/", func(r chi.Router) {
			r.Use(middleware2.Auth)
			r.Use(middleware2.GzipHandle)
			r.Post("/orders", h.SendOrderHandler)                         //загрузка пользователем номера заказа для расчёта
			r.Get("/orders", h.GetOrderHandler)                           //получение списка загруженных пользователем номеров заказов, статусов их обработки и информации о начислениях
			r.Get("/balance", h.GetBalanceHandler)                        //получение текущего баланса счёта баллов лояльности пользователя
			r.Post("/balance/withdraw", h.WithdrawHandler)                //запрос на списание баллов с накопительного счёта в счёт оплаты нового заказа
			r.Get("/balance/withdrawals", h.GetWithdrawalsHistoryHandler) //получение информации о выводе средств с накопительного счёта пользователем
		})
	})

	address := viper.GetString("RUN_ADDRESS")

	if err := http.ListenAndServe(address, r); err != nil {
		log.Fatal(err)
	}
}
