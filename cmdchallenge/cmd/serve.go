package main

import (
	"flag"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gitlab.com/jarv/cmdchallenge/internal/cmdserver"
	"gitlab.com/jarv/cmdchallenge/internal/config"
	"gitlab.com/jarv/cmdchallenge/internal/dashboard"
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
	genDashboard := flag.Bool("genDashboard", false, "create a dashboard snapshot")
	addr := flag.String("addr", "localhost:8181", "bind address")

	flag.Parse()

	log := logger.NewLogger()
	cfg := config.New()

	if *genDashboard {
		log.Info("Generating image for dashboard")
		// /d/9dMXL2N7z/cmd-application?kiosk&orgId=1
		err := dashboard.New(log).Capture()

		if err != nil {
			log.Panic(err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	router := mux.NewRouter()
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

	router.PathPrefix("/c/s").Handler(handlers.ProxyHeaders(sol.Handler()))
	router.PathPrefix("/c/r").Handler(handlers.ProxyHeaders(server.Handler()))
	router.Path("/metrics").Handler(handlers.ProxyHeaders(promhttp.Handler()))

	log.Info("Listening on " + *addr)

	srv := http.Server{
		Handler:      router,
		Addr:         *addr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}
