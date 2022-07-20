package solutions

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/didip/tollbooth"
	"github.com/sirupsen/logrus"
	"gitlab.com/jarv/cmdchallenge/internal/config"
	"gitlab.com/jarv/cmdchallenge/internal/metrics"
	"gitlab.com/jarv/cmdchallenge/internal/runner"
)

var (
	ErrInvalidMethod = errors.New("invalid method for solutions")
	ErrInvalidParam  = errors.New("invalid parameter for solutions")
	ErrStore         = errors.New("storage error for solutions")
)

const (
	maxRequestsSec = 2
)

type jsonCmds struct {
	Cmds []string `json:"cmds"`
}

type Solutions struct {
	log       *logrus.Logger
	config    *config.Config
	metrics   *metrics.Metrics
	store     runner.RunnerResultStorer
	rateLimit bool
}

func New(log *logrus.Logger, cfg *config.Config, m *metrics.Metrics, store runner.RunnerResultStorer, rateLimit bool) *Solutions {
	return &Solutions{
		log:       log,
		config:    cfg,
		metrics:   m,
		store:     store,
		rateLimit: rateLimit,
	}
}

func (s *Solutions) httpError(w http.ResponseWriter, e error, statusCode int) {
	s.metrics.CmdErrors.WithLabelValues(e.Error()).Inc()
	http.Error(w, e.Error(), statusCode)
}

func (s *Solutions) Handler() http.Handler {
	if s.rateLimit {
		lmt := tollbooth.NewLimiter(float64(maxRequestsSec), nil)
		s.log.Infof("Setting rate limit for solutions req/sec: %d", maxRequestsSec)
		lmt.SetIPLookups([]string{"RemoteAddr"})
		lmt.SetMessage("Your are sending requests too fast, slow down!")
		lmt.SetOnLimitReached(func(w http.ResponseWriter, r *http.Request) {
			s.log.Warn(fmt.Sprintf("Rate limit reached for `%s` on %s", r.RemoteAddr, r.RequestURI))
			s.metrics.ResponseStatus.WithLabelValues(strconv.Itoa(lmt.GetStatusCode()), r.RequestURI).Inc()
		})
		return tollbooth.LimitHandler(lmt, http.HandlerFunc(s.runHandler))
	} else {
		return http.HandlerFunc(s.runHandler)
	}
}

func (s *Solutions) runHandler(w http.ResponseWriter, req *http.Request) {
	// For local development allow alternate port for POST requests
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Cache-Control", "public, max-age=1800, s-maxage=1800")

	s.log.WithFields(logrus.Fields{
		"URI":  req.RequestURI,
		"Addr": req.RemoteAddr,
	}).Info("Solution request received")

	if req.Method != http.MethodGet {
		s.log.Errorf("expect GET, got %v", req.Method)
		s.httpError(w, ErrInvalidMethod, http.StatusMethodNotAllowed)
		return
	}

	slugs, ok := req.URL.Query()["slug"]
	if !ok || len(slugs[0]) < 1 {
		s.log.Error("Url Param 'slug' is missing")
		s.httpError(w, ErrInvalidParam, http.StatusInternalServerError)
		return
	}

	cmds, err := s.store.TopCmdsForSlug(slugs[0])
	if err != nil {
		s.log.Errorf("Unable to query top commands for %s", slugs[0])
		s.httpError(w, ErrStore, http.StatusInternalServerError)
		return
	}

	b, _ := json.Marshal(&jsonCmds{
		Cmds: cmds,
	})
	fmt.Fprint(w, string(b))
}
