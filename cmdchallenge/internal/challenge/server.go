package challenge

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"strconv"

	// "github.com/gdexlab/go-render/render"
	"github.com/didip/tollbooth/v7"
	"gitlab.com/jarv/cmdchallenge/internal/config"
	"gitlab.com/jarv/cmdchallenge/internal/metrics"
	"gitlab.com/jarv/cmdchallenge/internal/store"
)

const (
	MaxCMDLength         = 300
	maxServerRequestsSec = 0.5
	burst                = 2
)

type Server struct {
	log            *slog.Logger
	cfg            *config.Config
	metrics        *metrics.Metrics
	runnerExecutor RunnerExecutor
	cmdStorer      store.CmdStorer
}

type CmdResponse struct {
	Cached        *bool   `json:",omitempty"`
	Correct       *bool   `json:",omitempty"`
	Error         *string `json:",omitempty"` // Error string for failing checks
	ErrorInternal *string `json:",omitempty"` // Internal errors that will never be cached
	ExitCode      *int    `json:",omitempty"`
	Output        *string `json:",omitempty"`
}

func NewServer(
	log *slog.Logger,
	cfg *config.Config,
	m *metrics.Metrics,
	r RunnerExecutor,
	s store.CmdStorer,
) *Server {
	return &Server{
		log:            log,
		cfg:            cfg,
		metrics:        m,
		runnerExecutor: r,
		cmdStorer:      s,
	}
}

func (c *Server) httpError(w http.ResponseWriter, e error, statusCode int) {
	var chError *ChallengeError
	if errors.As(e, &chError) {
		c.metrics.CmdErrors.WithLabelValues(chError.Error(), chError.typ).Inc()
	} else {
		c.metrics.CmdErrors.WithLabelValues(e.Error(), TypeServer).Inc()
	}
	http.Error(w, e.Error(), statusCode)
}

func (c *Server) Handler() http.Handler {
	if c.cfg.RateLimit {
		lmt := tollbooth.NewLimiter(float64(maxServerRequestsSec), nil)
		c.log.Info(fmt.Sprintf("Setting rate limit req/sec: %f burst: %d", maxServerRequestsSec, burst))
		lmt.SetIPLookups([]string{"RemoteAddr"})
		lmt.SetBurst(burst)
		lmt.SetMessage("Your are sending command too fast, slow down!")
		lmt.SetOnLimitReached(func(w http.ResponseWriter, r *http.Request) {
			c.log.Info("Rate limit reached", "RemoteAddr", r.RemoteAddr, "RequestURI", r.RequestURI)
			c.metrics.ResponseStatus.WithLabelValues(strconv.Itoa(lmt.GetStatusCode()), r.RequestURI).Inc()
		})
		return tollbooth.LimitHandler(lmt, http.HandlerFunc(c.runHandler))
	} else {
		return http.HandlerFunc(c.runHandler)
	}
}

func (c *Server) runHandler(w http.ResponseWriter, req *http.Request) {
	// For local development allow alternate port for POST requests
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Cache-Control", "no-store, max-age=0")

	c.log.Info("Command request received",
		"URI", req.RequestURI,
		"Addr", req.RemoteAddr,
	)

	if req.Method != http.MethodPost {
		c.log.Error("expect POST", "method", req.Method)
		c.httpError(w, ErrServerInvalidMethod, http.StatusMethodNotAllowed)
		return
	}

	slug := req.PostFormValue("slug")
	cmd := req.PostFormValue("cmd")

	if err := isValidRequest(slug, cmd); err != nil {
		c.log.Error("Invalid request", "slug", slug, "cmd", cmd)
		c.httpError(w, err, http.StatusInternalServerError)
		return
	}

	cmd = decodeCmd(cmd)

	if len(cmd) > MaxCMDLength {
		c.log.Error("Command is too long", "len", len(cmd))
		c.httpError(w, ErrServerCmdTooLong, http.StatusForbidden)
		return
	}

	ch, err := NewChallenge(ChallengeOptions{Slug: slug})
	if err != nil {
		c.log.Error("Unable to parse challenge", "slug", slug)
		c.httpError(w, ErrServerInvalidChallenge, http.StatusInternalServerError)
		return
	}

	if slug != ch.Slug() {
		c.log.Error("Challenge slug doesn't match config", "slug", slug, "config", ch.Slug())
		c.httpError(w, ErrServerUnknown, http.StatusInternalServerError)
		return
	}

	c.log.Info("Got command",
		"cmd", cmd,
		"remoteAddr", req.RemoteAddr,
		"chDir", ch.Dir(),
		"slug", ch.Slug(),
	)

	jsonResp, err := c.runCmd(cmd, ch)
	if err != nil {
		c.httpError(w, err, http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, jsonResp)
}

func (c *Server) runAndStoreCmd(cmd string, ch *Challenge) (*store.CmdStore, error) {
	cmdResp, err := c.runnerExecutor.RunContainer(cmd, ch)
	if err == ErrRunnerTimeout {
		c.log.Error("Timeout running command", "err", err)
		return nil, &ChallengeError{msg: RunnerTimeout, typ: TypeRunner}
	}
	if err != nil {
		c.log.Error("Runner error", "err", err)
		return nil, &ChallengeError{msg: RunnerError, typ: TypeRunner}
	}

	if cmdResp.ErrorInternal != nil {
		return nil, &ChallengeError{msg: *cmdResp.ErrorInternal, typ: TypeRunCmd}
	}

	if cmdResp.Correct == nil || cmdResp.ExitCode == nil {
		c.log.Error("Invalid response from runner, `Correct`, `ExitCode` must be set for responses that aren't internal errors")
		return nil, &ChallengeError{msg: RunCmdInvalid, typ: TypeRunCmd}
	}

	cmdStore := &store.CmdStore{
		Cmd:      toPtr(cmd),
		Slug:     toPtr(ch.Slug()),
		Version:  toPtr(ch.Version()),
		Correct:  cmdResp.Correct,
		ExitCode: cmdResp.ExitCode,
		Output:   cmdResp.Output,
	}

	if cmdResp.Error != nil {
		cmdStore.Error = cmdResp.Error
	}

	if err = c.cmdStorer.CreateResult(cmdStore); err != nil {
		c.log.Error("Unable to create result", "err", err)
		return nil, &ChallengeError{msg: StoreError, typ: TypeStore}
	}

	return cmdStore, nil
}

func (c *Server) runCmd(cmd string, ch *Challenge) (string, error) {
	resultCached := true

	labels := metrics.CmdProcessedLabels{
		Slug:    ch.Slug(),
		Cached:  "true",
		Correct: "false",
	}

	cmdStore, err := c.cmdStorer.GetResult(cmd, ch.Slug(), ch.Version())
	if err == store.ErrResultNotFound {
		c.log.Info("No result found in cache, executing cmd",
			"cmd", cmd,
			"chDir", ch.Dir(),
			"slug", ch.Slug(),
		)
		// Run a new command and store it
		resultCached = false
		labels.Cached = "false"
		cmdStore, err = c.runAndStoreCmd(cmd, ch)
		if err != nil {
			return "", err
		}
	}

	if err != nil {
		c.log.Error("Unable to query result", "err", err)
		return "", &ChallengeError{msg: StoreQueryError, typ: TypeStore}
	}

	c.log.Info("Incrementing result", "cmd", cmd, "version", ch.Version())
	if err = c.cmdStorer.IncrementResult(cmd, ch.Slug(), ch.Version()); err != nil {
		c.log.Error("Unable to increment result counter", "err", err)
		return "", &ChallengeError{msg: StoreQueryError, typ: TypeStore}
	}

	resp := CmdResponse{
		Correct:  cmdStore.Correct,
		Error:    cmdStore.Error,
		ExitCode: cmdStore.ExitCode,
		Output:   cmdStore.Output,
	}

	resp.Cached = toPtr(resultCached)

	b, err := json.Marshal(resp)
	if err != nil {
		return "", ErrServerDecode
	}

	if *cmdStore.Correct {
		labels.Correct = "true"
	}

	c.metrics.CmdProcessed.WithLabelValues(labels.Cached, labels.Slug, labels.Correct).Inc()

	return string(b), nil
}

func isValidRequest(slug, cmd string) error {
	if slug == "" || cmd == "" {
		return ErrServerInvalidRequest
	}

	matched, _ := regexp.MatchString(`^\w+$`, slug)

	if !matched {
		return ErrServerInvalidChallenge
	}

	return nil
}

func decodeCmd(cmd string) string {
	// if the base64 decode fails assume the command was passed in
	// without encoding
	decoded, err := base64.StdEncoding.DecodeString(cmd)
	if err == nil {
		return string(decoded)
	}
	return cmd
}

type ptrConvert interface {
	string | bool | int
}

func toPtr[T ptrConvert](i T) *T {
	return &i
}
