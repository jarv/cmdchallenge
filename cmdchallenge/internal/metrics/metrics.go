package metrics

import (
	"database/sql"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type CmdProcessedLabels struct {
	Cached  string
	Slug    string
	Correct string
}

type Metrics struct {
	log            *slog.Logger
	CmdProcessed   *prometheus.CounterVec
	CmdErrors      *prometheus.CounterVec
	TotalRequests  *prometheus.CounterVec
	ResponseStatus *prometheus.CounterVec
	HTTPDuration   *prometheus.HistogramVec
}

var singleMetrics *Metrics

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func NewResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{w, http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func New(log *slog.Logger) *Metrics {
	if singleMetrics != nil {
		return singleMetrics
	}

	log.Info("Registering base metrics")
	m := Metrics{
		log: log,
		CmdProcessed: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "cmd_processed_total",
				Help: "The total number of processed cmds",
			},
			[]string{"cached", "slug", "correct"}),
		CmdErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "cmd_errors_total",
				Help: "The total number errors",
			},
			[]string{"error", "type"}),
		TotalRequests: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Number of get requests.",
			},
			[]string{"path"}),
		ResponseStatus: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "response_status_total",
				Help: "Status of HTTP response",
			},
			[]string{"status", "path"}),
		HTTPDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name: "http_response_time_seconds",
				Help: "Duration of HTTP requests.",
			},
			[]string{"status", "path"}),
	}

	singleMetrics = &m
	return singleMetrics
}

func (m *Metrics) DBStatsRegister(db *sql.DB, dbName string) {
	m.log.Info("Registering DB stats")
	prometheus.MustRegister(collectors.NewDBStatsCollector(db, dbName))
}

func (m *Metrics) PrometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		route := mux.CurrentRoute(r)
		path, _ := route.GetPathTemplate()

		timer := prometheus.NewTimer(nil)
		rw := NewResponseWriter(w)
		next.ServeHTTP(rw, r)

		statusCode := rw.statusCode

		m.ResponseStatus.WithLabelValues(strconv.Itoa(statusCode), path).Inc()
		m.TotalRequests.WithLabelValues(path).Inc()

		d := timer.ObserveDuration()
		m.HTTPDuration.WithLabelValues(strconv.Itoa(statusCode), path).Observe(d.Seconds())
	})
}
