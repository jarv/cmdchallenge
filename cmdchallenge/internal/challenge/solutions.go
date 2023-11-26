package challenge

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/didip/tollbooth/v7"
	"gitlab.com/jarv/cmdchallenge/internal/config"
	"gitlab.com/jarv/cmdchallenge/internal/metrics"
	"gitlab.com/jarv/cmdchallenge/internal/store"
)

const (
	maxSolutionsRequestsSec = 2
)

type jsonCmds struct {
	Cmds []string `json:"cmds"`
}

type Solutions struct {
	log       *slog.Logger
	cfg       *config.Config
	metrics   *metrics.Metrics
	cmdStorer store.CmdStorer
	rateLimit bool
}

func NewSolutions(log *slog.Logger, cfg *config.Config, m *metrics.Metrics, s store.CmdStorer) *Solutions {
	return &Solutions{
		log:       log,
		cfg:       cfg,
		metrics:   m,
		cmdStorer: s,
		rateLimit: cfg.RateLimit,
	}
}

func (s *Solutions) httpError(w http.ResponseWriter, e error, statusCode int) {
	s.metrics.CmdErrors.WithLabelValues(e.Error()).Inc()
	http.Error(w, e.Error(), statusCode)
}

func (s *Solutions) Handler() http.Handler {
	if s.rateLimit {
		lmt := tollbooth.NewLimiter(float64(maxSolutionsRequestsSec), nil)
		s.log.Info("Setting rate limit for solutions req/sec", "maxRequestsSec", maxSolutionsRequestsSec)
		lmt.SetIPLookups([]string{"RemoteAddr"})
		lmt.SetMessage("Your are sending requests too fast, slow down!")
		lmt.SetOnLimitReached(func(w http.ResponseWriter, r *http.Request) {
			s.log.Info("Rate limit reached", "RemoteAddr", r.RemoteAddr, "RequestURI", r.RequestURI)
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

	s.log.Info("Solution request received", "URI", req.RequestURI, "Addr", req.RemoteAddr)

	if req.Method != http.MethodGet {
		s.log.Error("expected GET", "method", req.Method)
		s.httpError(w, ErrSolutionsInvalidMethod, http.StatusMethodNotAllowed)
		return
	}

	slugs, ok := req.URL.Query()["slug"]
	if !ok || len(slugs[0]) < 1 {
		s.log.Error("Url Param 'slug' is missing")
		s.httpError(w, ErrSolutionsInvalidParam, http.StatusInternalServerError)
		return
	}

	cmds, err := s.cmdStorer.TopCmdsForSlug(slugs[0])
	if err != nil {
		s.log.Error("Unable to query top commands", "slug", slugs[0], "err", err)
		s.httpError(w, ErrSolutionsStore, http.StatusInternalServerError)
		return
	}

	b, _ := json.Marshal(&jsonCmds{
		Cmds: cmds,
	})
	fmt.Fprint(w, string(b))
}
