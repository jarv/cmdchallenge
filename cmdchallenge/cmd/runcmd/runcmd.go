package main

import (
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	//nolint:gosec,G108
	_ "net/http/pprof"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gitlab.com/jarv/cmdchallenge/internal/challenge"
	"gitlab.com/jarv/cmdchallenge/internal/config"
	"gitlab.com/jarv/cmdchallenge/internal/metrics"
	"gitlab.com/jarv/cmdchallenge/internal/runcmd"
	"gitlab.com/jarv/cmdchallenge/internal/store"
)

func handleCmd(log *slog.Logger, slug string, cfg *config.Config) error {
	if slug == "" {
		return errors.New("you must provide a slug name for the command runner")
	}

	if flag.NArg() != 1 {
		return errors.New("you must specificy a command to run")
	}

	ch, err := challenge.NewChallenge(challenge.ChallengeOptions{Slug: slug})
	if err != nil {
		return errors.New("Unable to parse challenge: " + err.Error())
	}

	decoded, err := base64.StdEncoding.DecodeString(flag.Args()[0])
	var command string
	if err != nil {
		command = flag.Args()[0]
	} else {
		command = string(decoded)
	}
	fmt.Println(runcmd.New(log, cfg).Run(ch, command))

	return nil
}

func handleServer(log *slog.Logger, cfg *config.Config, addr string) {
	cmdMetrics := metrics.New(log)
	router := mux.NewRouter()
	runner := challenge.NewRunner(log, cfg)

	var cmdStorer store.CmdStorer
	var err error
	if cfg.DevMode {
		cmdStorer, err = store.NewMemStore()
	} else {
		cmdStorer, err = store.NewSQLStore(log, cmdMetrics, cfg.DBFile)
	}
	if err != nil {
		log.Error("Unable to initialize db!", "err", err)
		return
	}

	solutions := challenge.NewSolutions(log, cfg, cmdMetrics, cmdStorer)
	server := challenge.NewServer(log, cfg, cmdMetrics, runner, cmdStorer)

	router.Use(cmdMetrics.PrometheusMiddleware)
	router.PathPrefix("/c/s").Handler(handlers.ProxyHeaders(solutions.Handler()))
	router.PathPrefix("/c/r").Handler(handlers.ProxyHeaders(server.Handler()))
	router.Path("/metrics").Handler(handlers.ProxyHeaders(promhttp.Handler()))
	router.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)
	router.PathPrefix("/").Handler(http.FileServer(http.Dir(cfg.StaticDistDir)))

	log.Info("Listening on " + addr)

	srv := http.Server{
		Handler:           router,
		Addr:              addr,
		WriteTimeout:      15 * time.Second,
		ReadTimeout:       15 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Error("Failed to setup listener!", "err", err)
	}
}

func main() {
	devMode := flag.Bool("dev", lookupEnvOrVal("CMD_DEV_MODE", false), "run in development mode")
	rateLimit := flag.Bool("setRateLimit", lookupEnvOrVal("CMD_SET_RATE_LIMIT", false), "set rate limits")
	devTag := flag.Bool("devTag", lookupEnvOrVal("CMD_DEV_TAG", false), "use a dev tag for container images")
	dbFile := flag.String("dbFile", lookupEnvOrVal("CMD_DB_FILE", "/app/db.sqlite3"), "path to the db file")
	staticDistDir := flag.String("staticDistDir", lookupEnvOrVal("CMD_STATIC_DIST_DIR", "/app/dist"), "path to static files")
	cmd := flag.Bool("cmd", false, "execute a command inside the runner")
	slug := flag.String("slug", "", "slug for the command executor")
	addr := flag.String("addr", ":8181", "bind address")

	flag.Parse()

	log := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	cfg := config.New(config.ConfigOpts{
		DevMode:       *devMode,
		RateLimit:     *rateLimit,
		DevTag:        *devTag,
		DBFile:        *dbFile,
		StaticDistDir: *staticDistDir,
	})

	if *cmd {
		if err := handleCmd(log, *slug, cfg); err != nil {
			log.Error("Command failed", "err", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	log.Info(fmt.Sprintf("%#v", cfg))
	handleServer(log, cfg, *addr)
}

type envLookup interface {
	string | int | bool
}

func lookupEnvOrVal[T envLookup](key string, defaultVal T) T {
	if val, ok := os.LookupEnv(key); ok {
		var ret T
		switch p := any(&ret).(type) {
		case *string:
			*p = val
		case *int:
			*p, _ = strconv.Atoi(val)
		case *bool:
			*p = true
		}
		return ret
	}

	return defaultVal
}
