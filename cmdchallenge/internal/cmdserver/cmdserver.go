package cmdserver

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/http"
	"path"
	"regexp"
	"strconv"

	// "github.com/gdexlab/go-render/render"
	"github.com/didip/tollbooth"
	"github.com/sirupsen/logrus"
	"gitlab.com/jarv/cmdchallenge/internal/challenge"
	"gitlab.com/jarv/cmdchallenge/internal/config"
	"gitlab.com/jarv/cmdchallenge/internal/metrics"
	"gitlab.com/jarv/cmdchallenge/internal/runner"
	"k8s.io/client-go/1.5/pkg/util/json"
)

const (
	MaxCMDLength   = 300
	maxRequestsSec = 0.5
	burst          = 2
)

var (
	ErrInvalidSourceIP  = errors.New("unable to determine source IP")
	ErrCmdTooLong       = errors.New("command is too long")
	ErrInvalidMethod    = errors.New("invalid method")
	ErrInvalidRequest   = errors.New("request must include slug and cmd")
	ErrInvalidChallenge = errors.New("invalid challenge")
	ErrRunner           = errors.New("runner error")
	ErrTimeout          = errors.New("command timed out")
	ErrUnknown          = errors.New("unknown error")
	ErrDecode           = errors.New("decode error")
	ErrStore            = errors.New("storage error")
)

type CmdServer struct {
	log       *logrus.Logger
	config    *config.Config
	metrics   *metrics.Metrics
	runner    runner.RunnerExecutor
	store     runner.RunnerResultStorer
	rateLimit bool
}

type CmdResponse struct {
	Cached   *bool   `json:",omitempty"`
	Correct  *bool   `json:",omitempty"`
	Error    *string `json:",omitempty"`
	ExitCode *int32  `json:",omitempty"`
	Output   *string `json:",omitempty"`
}

func New(
	logger *logrus.Logger,
	cfg *config.Config,
	m *metrics.Metrics,
	run runner.RunnerExecutor,
	resultStore runner.RunnerResultStorer,
	rateLimit bool,
) *CmdServer {
	return &CmdServer{
		log:       logger,
		config:    cfg,
		metrics:   m,
		runner:    run,
		store:     resultStore,
		rateLimit: rateLimit,
	}
}

func (c *CmdServer) httpError(w http.ResponseWriter, e error, statusCode int) {
	c.metrics.CmdErrors.WithLabelValues(e.Error()).Inc()
	http.Error(w, e.Error(), statusCode)
}

func (c *CmdServer) Handler() http.Handler {
	if c.rateLimit {
		lmt := tollbooth.NewLimiter(float64(maxRequestsSec), nil)
		c.log.Info(fmt.Sprintf("Setting rate limit req/sec: %f burst: %d", maxRequestsSec, burst))
		lmt.SetIPLookups([]string{"RemoteAddr", "X-Forwarded-For", "X-Real-IP"})
		lmt.SetBurst(burst)
		lmt.SetMessage("Your are sending command too fast, slow down!")
		lmt.SetOnLimitReached(func(w http.ResponseWriter, r *http.Request) {
			c.log.Warn(fmt.Sprintf("Rate limit reached for `%s` on %s", r.RemoteAddr, r.RequestURI))
			c.metrics.ResponseStatus.WithLabelValues(strconv.Itoa(lmt.GetStatusCode()), r.RequestURI).Inc()
		})
		return tollbooth.LimitHandler(lmt, http.HandlerFunc(c.runHandler))
	} else {
		return http.HandlerFunc(c.runHandler)
	}
}

func (c *CmdServer) runHandler(w http.ResponseWriter, req *http.Request) {
	// For local development allow alternate port for POST requests
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Cache-Control", "no-store, max-age=0")

	if req.Method != http.MethodPost {
		c.log.Errorf("expect POST, got %v", req.Method)
		c.httpError(w, ErrInvalidMethod, http.StatusMethodNotAllowed)
		return
	}

	slug := req.PostFormValue("slug")
	cmd := req.PostFormValue("cmd")

	if err := isValidRequest(slug, cmd); err != nil {
		c.httpError(w, err, http.StatusInternalServerError)
		return
	}

	cmd = decodeCmd(cmd)

	if len(cmd) > MaxCMDLength {
		c.log.Error(fmt.Sprintf("Command is too long: %d", len(cmd)))
		c.httpError(w, ErrCmdTooLong, http.StatusForbidden)
		return
	}

	host, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		c.log.Error(fmt.Sprintf("Unable to determine source IP for `%s` : %s", req.RemoteAddr, err.Error()))
		c.httpError(w, ErrInvalidSourceIP, http.StatusInternalServerError)
		return
	}

	chFile := path.Join(c.config.ROVolumeDir, "ch", slug+".json")
	ch, err := challenge.New(chFile)
	if err != nil {
		c.log.Error("Unable to parse challenge: " + err.Error())
		c.httpError(w, ErrInvalidChallenge, http.StatusInternalServerError)
		return
	}

	if slug != ch.Slug() {
		c.log.Error(fmt.Sprintf("Challenge slug `%s` doesn't match config `%s`", slug, ch.Slug()))
		c.httpError(w, ErrUnknown, http.StatusInternalServerError)
		return
	}

	fingerprint, err := ch.Fingerprint(cmd)
	if err != nil {
		c.log.Error("Unable to generate command fingerprint: " + err.Error())
		c.httpError(w, ErrUnknown, http.StatusInternalServerError)
		return
	}

	jsonResp, err := c.runCmd(cmd, fingerprint, host, ch)
	if err != nil {
		c.log.Error(err.Error())
		c.httpError(w, err, http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, jsonResp)
}

func (c *CmdServer) runCmd(cmd, fingerprint, host string, ch *challenge.Challenge) (string, error) {
	c.log.WithFields(logrus.Fields{
		"cmd":         cmd,
		"fingerprint": fingerprint,
		"host":        host,
		"chDir":       ch.Dir(),
		"slug":        ch.Slug(),
	}).Info("Initiating command")

	result, err := c.store.GetResult(fingerprint)
	resultCached := true
	labels := metrics.CmdProcessedLabels{
		Slug:    ch.Slug(),
		Cached:  "true",
		Correct: "false",
	}

	if err == runner.ErrResultNotFound {
		resultCached = false
		labels.Cached = "false"
		result, err = c.runner.RunContainer(ch, cmd)
		if err == runner.ErrTimeout {
			return "", ErrTimeout
		}
		if err != nil {
			c.log.Errorln(err)
			return "", ErrRunner
		}

		if err = c.store.CreateResult(fingerprint, cmd, ch.Slug(), ch.Version(), result); err != nil {
			c.log.Errorf("Unable to create result: %s", err.Error())
			return "", ErrStore
		}
	}

	if err != nil {
		c.log.Errorf("Unable to query result: %s", err.Error())
		return "", ErrStore
	}

	if err = c.store.IncrementResult(fingerprint); err != nil {
		c.log.Errorf("Unable to increment result counter: %s", err.Error())
		return "", ErrStore
	}

	if result.Correct == nil || result.ExitCode == nil {
		c.log.Error("Invalid response, both `Correct` and `ExitCode` expected from runner!")
		return "", ErrUnknown
	}

	resp := CmdResponse{
		Correct:  result.Correct,
		Error:    result.Error,
		ExitCode: result.ExitCode,
		Output:   result.Output,
	}

	resp.Cached = boolPtr(resultCached)

	b, err := json.Marshal(resp)
	if err != nil {
		return "", ErrDecode
	}

	if *result.Correct {
		labels.Correct = "true"
	}

	c.metrics.CmdProcessed.WithLabelValues(labels.Cached, labels.Slug, labels.Correct).Inc()

	return string(b), nil
}

func isValidRequest(slug, cmd string) error {
	if slug == "" || cmd == "" {
		return ErrInvalidRequest
	}

	matched, _ := regexp.MatchString(`^\w+$`, slug)

	if !matched {
		return ErrInvalidChallenge
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

func intPtr(i int32) *int32 {
	return &i
}

func boolPtr(b bool) *bool {
	return &b
}

func strPtr(s string) *string {
	return &s
}
