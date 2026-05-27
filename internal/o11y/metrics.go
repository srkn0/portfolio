package o11y

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const unknownRoute = "unknown"

// Recorder is the surface the app uses to push numbers somewhere.
// Real impl: Prometheus client_golang with a Registry.
// Test impl: a spy that just remembers what was called.
type Recorder interface {
	IncRequest(method, route string, status int)
	ObserveRequestDuration(method, route string, status int, seconds float64)
	SetPostCount(n int)
	SetProjectCount(n int)
	SetBuildInfo(version, commit string)
}

// NewRecorder builds a Prometheus-backed Recorder and registers its collectors
// on the given registry. Standard Go runtime + process collectors are added
// alongside so memory, goroutines and GC are scraped automatically.
func NewRecorder(reg prometheus.Registerer) Recorder {
	r := &promRecorder{
		requests: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total HTTP requests received, labelled by method, route pattern and status.",
			},
			[]string{"method", "route", "status"},
		),
		duration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request duration in seconds.",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "route", "status"},
		),
		posts: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "portfolio_post_count",
			Help: "Number of blog posts loaded into the in-memory store.",
		}),
		projects: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "portfolio_project_count",
			Help: "Number of projects loaded into the in-memory store.",
		}),
		buildInfo: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "portfolio_build_info",
				Help: "Build info gauge; value is always 1, version and commit live on labels.",
			},
			[]string{"version", "commit"},
		),
	}
	reg.MustRegister(
		r.requests, r.duration, r.posts, r.projects, r.buildInfo,
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)
	return r
}

type promRecorder struct {
	requests  *prometheus.CounterVec
	duration  *prometheus.HistogramVec
	posts     prometheus.Gauge
	projects  prometheus.Gauge
	buildInfo *prometheus.GaugeVec
}

func (r *promRecorder) IncRequest(method, route string, status int) {
	r.requests.WithLabelValues(method, route, strconv.Itoa(status)).Inc()
}

func (r *promRecorder) ObserveRequestDuration(method, route string, status int, seconds float64) {
	r.duration.WithLabelValues(method, route, strconv.Itoa(status)).Observe(seconds)
}

func (r *promRecorder) SetPostCount(n int)    { r.posts.Set(float64(n)) }
func (r *promRecorder) SetProjectCount(n int) { r.projects.Set(float64(n)) }
func (r *promRecorder) SetBuildInfo(version, commit string) {
	r.buildInfo.Reset()
	r.buildInfo.WithLabelValues(version, commit).Set(1)
}

// Middleware wraps next and pushes one counter + one duration sample per request.
// Route pattern comes from RouteFromContext(r); when absent (e.g. a 404 before
// routing matched) a bounded constant "unknown" is used so cardinality stays small.
func Middleware(rec Recorder, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sr := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sr, r)

		route := RouteFromContext(r)
		if route == "" {
			route = unknownRoute
		}

		rec.IncRequest(r.Method, route, sr.status)
		rec.ObserveRequestDuration(r.Method, route, sr.status, time.Since(start).Seconds())
	})
}

// MetricsMiddleware returns chi-style middleware for the Recorder.
// Convenience wrapper around Middleware so server code can do r.Use(o11y.MetricsMiddleware(rec)).
func MetricsMiddleware(rec Recorder) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return Middleware(rec, next)
	}
}

// RouteFromContext extracts the matched chi route pattern. Returns "" when
// no chi route matched (e.g. requests not routed through chi, or pre-routing).
func RouteFromContext(r *http.Request) string {
	rctx := chi.RouteContext(r.Context())
	if rctx == nil {
		return ""
	}
	return rctx.RoutePattern()
}

// statusRecorder wraps an http.ResponseWriter to capture the status code
// for logging and metrics. Defaults to 200 because handlers that return
// without explicitly calling WriteHeader implicitly send 200.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

// defaultRegistry is the package-level Prometheus registry that MetricsHandler serves.
// Wiring code should NewRecorder(o11y.DefaultRegistry()) and then mount MetricsHandler().
var defaultRegistry = prometheus.NewRegistry()

// DefaultRegistry exposes the package-level Prometheus registry.
func DefaultRegistry() *prometheus.Registry { return defaultRegistry }

// MetricsHandler serves the metrics in Prometheus text exposition format
// from the package-level default registry.
func MetricsHandler() http.Handler {
	return MetricsHandlerFor(defaultRegistry)
}

// MetricsHandlerFor serves the metrics from the given registry.
// Useful in tests so registrations stay isolated.
func MetricsHandlerFor(reg *prometheus.Registry) http.Handler {
	return promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg})
}
