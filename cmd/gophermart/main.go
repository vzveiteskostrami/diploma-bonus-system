package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi"
	"github.com/vzveiteskostrami/diploma-bonus-system/internal/auth"
	"github.com/vzveiteskostrami/diploma-bonus-system/internal/checkout"
	"github.com/vzveiteskostrami/diploma-bonus-system/internal/compressing"
	"github.com/vzveiteskostrami/diploma-bonus-system/internal/config"
	"github.com/vzveiteskostrami/diploma-bonus-system/internal/dbf"
	"github.com/vzveiteskostrami/diploma-bonus-system/internal/logging"
	"github.com/vzveiteskostrami/diploma-bonus-system/internal/routes"
)

var (
	srv *http.Server
)

func main() {
	logging.LoggingInit()
	defer logging.LoggingSync()

	config.ReadData()

	dbf.Init()
	defer dbf.Close()

	srv = &http.Server{
		Addr:        config.Addresses.In.Host + ":" + strconv.Itoa(config.Addresses.In.Port),
		Handler:     mainRouter(),
		IdleTimeout: time.Second * 1,
	}

	logging.S().Infow(
		"Starting server",
		"addr", config.Addresses.In.Host+":"+strconv.Itoa(config.Addresses.In.Port),
	)
	logging.S().Fatal(srv.ListenAndServe())
}

func mainRouter() chi.Router {
	r := chi.NewRouter()

	r.Route("/api/user/register", func(r chi.Router) {
		r.Use(compressing.GZIPHandle)
		r.Use(logging.WithLogging)
		r.Post("/", routes.Registerf)
	})

	r.Route("/api/user/login", func(r chi.Router) {
		r.Use(compressing.GZIPHandle)
		r.Use(logging.WithLogging)
		r.Post("/", routes.Authf)
	})

	r.Route("/api/user", func(r chi.Router) {
		r.Use(compressing.GZIPHandle)
		r.Use(logging.WithLogging)
		r.Use(auth.AuthHandle)
		r.Use(checkout.RunCheck)
		r.Post("/orders", routes.OrdersPostf)
		r.Get("/orders", routes.OrdersGetf)
		r.Get("/balance", routes.BalanceGetf)
		r.Post("/balance/withdraw", routes.WithdrawPostf)
		r.Get("/withdrawals", routes.WithdrawGetf)
	})

	return r
}
