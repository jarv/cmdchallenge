package main

import (
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"strconv"

	//nolint:gosec,G108
	_ "net/http/pprof"
	"os"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/zerologr"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"gitlab.com/jarv/cmdchallenge/internal/challenge"
	"gitlab.com/jarv/cmdchallenge/internal/config"
	"gitlab.com/jarv/cmdchallenge/internal/metrics"
	"gitlab.com/jarv/cmdchallenge/internal/runcmd"
	"gitlab.com/jarv/cmdchallenge/internal/store"
)

func handleCmd(log logr.Logger, slug string, cfg *config.Config) error {
	if slug == "" {
		return errors.New("you must provide a slug name for the command runner")
	}

	if flag.NArg() != 1 {
		return errors.New("you must specificy a command to run")
	}

	chJSON, err := cfg.JSONForSlug(slug)
	if err != nil {
		log.Error(err, "Unable to open challenge file for slug", "slug", slug)
		return err
	}

	ch, err := challenge.NewChallenge(chJSON)
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

func handleServer(log logr.Logger, cfg *config.Config, addr string, devMode, setRateLimit bool) {
	cmdMetrics := metrics.New(log)
	router := mux.NewRouter()
	runner := challenge.NewRunner(log, cfg)

	var cmdStorer store.CmdStorer
	var err error
	if devMode {
		cmdStorer, err = store.NewMemStore()
	} else {
		cmdStorer, err = store.NewSQLStore(log, cmdMetrics, cfg.SQLiteDBFile)
	}
	if err != nil {
		panic(err)
	}

	solutions := challenge.NewSolutions(log, cfg, cmdMetrics, cmdStorer, setRateLimit)
	server := challenge.NewServer(log, cfg, cmdMetrics, runner, cmdStorer, setRateLimit)

	router.Use(cmdMetrics.PrometheusMiddleware)
	router.PathPrefix("/c/s").Handler(handlers.ProxyHeaders(solutions.Handler()))
	router.PathPrefix("/c/r").Handler(handlers.ProxyHeaders(server.Handler()))
	router.Path("/metrics").Handler(handlers.ProxyHeaders(promhttp.Handler()))
	router.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)

	log.Info("Listening on " + addr)

	srv := http.Server{
		Handler:           router,
		Addr:              addr,
		WriteTimeout:      15 * time.Second,
		ReadTimeout:       15 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
	}

	log.Error(srv.ListenAndServe(), "Finished")
}

func newLogger(color bool) logr.Logger {
	output := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339, NoColor: !color}
	zl := zerolog.New(output).With().Caller().Timestamp().Logger()
	zerologr.VerbosityFieldName = ""

	zerolog.TimestampFunc = func() time.Time {
		return time.Now().UTC()
	}
	zerolog.CallerMarshalFunc = func(_ uintptr, file string, line int) string {
		short := file
		for i := len(file) - 1; i > 0; i-- {
			if file[i] == '/' {
				short = file[i+1:]
				break
			}
		}
		file = short
		return file + ":" + strconv.Itoa(line)
	}

	return zerologr.New(&zl)
}

func main() {
	devMode := flag.Bool("dev", false, "run in development mode")
	setRateLimit := flag.Bool("setRateLimit", false, "set rate limits")
	cmd := flag.Bool("cmd", false, "execute a command inside the runner")
	slug := flag.String("slug", "", "slug for the command executor")
	addr := flag.String("addr", "localhost:8181", "bind address")

	flag.Parse()

	log := newLogger(true)
	logNoColor := newLogger(false)
	cfg := config.New()

	if *cmd {
		if err := handleCmd(logNoColor, *slug, cfg); err != nil {
			log.Error(err, "Command failed")
			os.Exit(1)
		}
		os.Exit(0)
	}

	handleServer(log, cfg, *addr, *devMode, *setRateLimit)
}
