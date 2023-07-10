package main

import (
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	//nolint:gosec,G108
	_ "net/http/pprof"

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

func handleServer(log logr.Logger, cfg *config.Config, addr string) {
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
		log.Error(err, "Unable to initialize db!")
		return
	}

	solutions := challenge.NewSolutions(log, cfg, cmdMetrics, cmdStorer)
	server := challenge.NewServer(log, cfg, cmdMetrics, runner, cmdStorer)

	router.Use(cmdMetrics.PrometheusMiddleware)
	router.PathPrefix("/c/s").Handler(handlers.ProxyHeaders(solutions.Handler()))
	router.PathPrefix("/c/r").Handler(handlers.ProxyHeaders(server.Handler()))
	router.Path("/metrics").Handler(handlers.ProxyHeaders(promhttp.Handler()))
	router.PathPrefix("/").Handler(http.FileServer(http.Dir(cfg.StaticDistDir)))
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
	devMode := flag.Bool("dev", lookupEnvOrVal("CMD_DEV_MODE", false), "run in development mode")
	rateLimit := flag.Bool("setRateLimit", lookupEnvOrVal("CMD_SET_RATE_LIMIT", false), "set rate limits")
	devTag := flag.Bool("devTag", lookupEnvOrVal("CMD_DEV_TAG", false), "use a dev tag for container images")
	dbFile := flag.String("dbFile", lookupEnvOrVal("CMD_DB_FILE", "/app/db.sqlite3"), "path to the db file")
	staticDistDir := flag.String("staticDistDir", lookupEnvOrVal("CMD_STATIC_DIST_DIR", "/app/dist"), "path to static files")
	cmd := flag.Bool("cmd", false, "execute a command inside the runner")
	slug := flag.String("slug", "", "slug for the command executor")
	addr := flag.String("addr", ":8181", "bind address")

	flag.Parse()

	log := newLogger(true)
	logNoColor := newLogger(false)
	cfg := config.New(config.ConfigOpts{
		DevMode:       *devMode,
		RateLimit:     *rateLimit,
		DevTag:        *devTag,
		DBFile:        *dbFile,
		StaticDistDir: *staticDistDir,
	})

	if *cmd {
		if err := handleCmd(logNoColor, *slug, cfg); err != nil {
			log.Error(err, "Command failed")
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
