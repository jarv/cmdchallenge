package main

import (
	"flag"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gitlab.com/jarv/cmdchallenge/internal/cmdserver"
	"gitlab.com/jarv/cmdchallenge/internal/config"
	"gitlab.com/jarv/cmdchallenge/internal/logger"
	"gitlab.com/jarv/cmdchallenge/internal/memstore"
	"gitlab.com/jarv/cmdchallenge/internal/metrics"
	"gitlab.com/jarv/cmdchallenge/internal/runner"
	"gitlab.com/jarv/cmdchallenge/internal/solutions"
	"gitlab.com/jarv/cmdchallenge/internal/sqlstore"
)

func main() {
	devMode := flag.Bool("dev", false, "run in development mode")
	rateLimit := flag.Bool("rateLimit", false, "set rate limits")
	addr := flag.String("addr", "localhost:8181", "bind address")

	flag.Parse()

	log := logger.NewLogger()

	router := mux.NewRouter()

	cfg := config.New()
	run := runner.New(log, cfg)
	cmdMetrics := metrics.New()

	var store runner.RunnerResultStorer
	var err error
	if *devMode {
		store, err = memstore.New()
	} else {
		store, err = sqlstore.New(cfg.SQLiteDBFile)
	}
	if err != nil {
		log.Panic(err)
	}

	sol := solutions.New(log, cfg, cmdMetrics, store, *rateLimit)
	server := cmdserver.New(log, cfg, cmdMetrics, run, store, *rateLimit)

	router.Use(cmdMetrics.PrometheusMiddleware)
	router.PathPrefix("/c/s").Handler(sol.Handler())
	router.PathPrefix("/c/r").Handler(server.Handler())
	router.Path("/metrics").Handler(promhttp.Handler())

	log.Info("Listening on " + *addr)

	srv := http.Server{
		Handler:      router,
		Addr:         *addr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}
