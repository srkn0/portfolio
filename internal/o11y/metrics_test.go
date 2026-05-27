package o11y_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/srkn0/main/internal/o11y"
)

// spy records every call the middleware makes against the Recorder.
type spy struct {
	incs      []incCall
	durations []durCall
	posts     int
	projects  int
	build     buildCall
}

type incCall struct {
	method string
	route  string
	status int
}

type durCall struct {
	method  string
	route   string
	status  int
	seconds float64
}

type buildCall struct {
	version string
	commit  string
}

func (s *spy) IncRequest(method, route string, status int) {
	s.incs = append(s.incs, incCall{method, route, status})
}
func (s *spy) ObserveRequestDuration(method, route string, status int, seconds float64) {
	s.durations = append(s.durations, durCall{method, route, status, seconds})
}
func (s *spy) SetPostCount(n int)    { s.posts = n }
func (s *spy) SetProjectCount(n int) { s.projects = n }
func (s *spy) SetBuildInfo(version, commit string) {
	s.build = buildCall{version, commit}
}

func TestMiddleware_incrementsRequestCounter(t *testing.T) {
	rec := &spy{}
	h := o11y.Middleware(rec, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/blog", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)

	if len(rec.incs) != 1 {
		t.Fatalf("expected 1 IncRequest call, got %d", len(rec.incs))
	}
	call := rec.incs[0]
	if call.method != http.MethodGet {
		t.Errorf("method = %q, want GET", call.method)
	}
	if call.status != http.StatusOK {
		t.Errorf("status = %d, want 200", call.status)
	}
}

func TestMiddleware_capturesNon200Status(t *testing.T) {
	rec := &spy{}
	h := o11y.Middleware(rec, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.NotFound(w, nil)
	}))

	req := httptest.NewRequest(http.MethodGet, "/blog/missing", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)

	if len(rec.incs) != 1 {
		t.Fatalf("expected 1 IncRequest call, got %d", len(rec.incs))
	}
	if rec.incs[0].status != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.incs[0].status)
	}
}

func TestMiddleware_observesDuration(t *testing.T) {
	rec := &spy{}
	h := o11y.Middleware(rec, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)

	if len(rec.durations) != 1 {
		t.Fatalf("expected 1 ObserveRequestDuration call, got %d", len(rec.durations))
	}
	if rec.durations[0].seconds < 0 {
		t.Errorf("seconds = %v, want >= 0", rec.durations[0].seconds)
	}
}

func TestMiddleware_usesRoutePatternNotRawURL(t *testing.T) {
	// Cardinality: /blog/post-a and /blog/post-b must collapse to /blog/{slug}.
	// This test only asserts that whatever ends up as the "route" label is
	// the same for both calls, which is enough to verify the middleware
	// doesn't naively use r.URL.Path.
	rec := &spy{}
	h := o11y.Middleware(rec, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for _, slug := range []string{"post-a", "post-b"} {
		req := httptest.NewRequest(http.MethodGet, "/blog/"+slug, nil)
		h.ServeHTTP(httptest.NewRecorder(), req)
	}

	if len(rec.incs) != 2 {
		t.Fatalf("expected 2 calls, got %d", len(rec.incs))
	}
	if rec.incs[0].route != rec.incs[1].route {
		t.Errorf("route labels differ for same pattern: %q vs %q (cardinality bomb)",
			rec.incs[0].route, rec.incs[1].route)
	}
}

func TestMetricsHandler_servesPrometheusFormat(t *testing.T) {
	// Prometheus's scraper requires a text/plain Content-Type and 200.
	// If either is wrong, the scrape silently fails and you see no metrics.
	rec := httptest.NewRecorder()
	o11y.MetricsHandler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	ct := rec.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "text/plain") && !strings.Contains(ct, "openmetrics") {
		t.Errorf("Content-Type = %q, want text/plain... or openmetrics", ct)
	}
}

func TestMetricsHandler_exposesAppMetrics(t *testing.T) {
	// After recording some activity through the real Recorder, the metrics
	// endpoint should expose the well-known names so Prometheus / Grafana
	// queries are stable.
	t.Skip("requires concrete prometheus-backed Recorder; enable once wired")

	// Pseudocode (uncomment once Recorder + Handler are real):
	//
	// reg := prometheus.NewRegistry()
	// rec := o11y.NewRecorder(reg)
	// rec.IncRequest("GET", "/blog", 200)
	// rec.SetPostCount(12)
	// rec.SetBuildInfo("v0.1.0", "abc1234")
	//
	// w := httptest.NewRecorder()
	// o11y.MetricsHandlerFor(reg).ServeHTTP(w, httptest.NewRequest("GET", "/metrics", nil))
	//
	// body := w.Body.String()
	// for _, want := range []string{
	//     "http_requests_total",
	//     "http_request_duration_seconds",
	//     "portfolio_post_count",
	//     "portfolio_build_info",
	// } {
	//     if !strings.Contains(body, want) {
	//         t.Errorf("metrics output missing %q", want)
	//     }
	// }
}

func TestMiddleware_unknownRouteFallsBackToBoundedLabel(t *testing.T) {
	// When no route is in context (e.g. before chi routed, or 404 on a
	// path that didn't match any route), the label must be a constant
	// like "unknown" instead of the raw path.
	rec := &spy{}
	h := o11y.Middleware(rec, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.NotFound(w, nil)
	}))

	for _, path := range []string{"/random-1", "/random-2", "/random-3"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		h.ServeHTTP(httptest.NewRecorder(), req)
	}

	if len(rec.incs) != 3 {
		t.Fatalf("expected 3 calls, got %d", len(rec.incs))
	}
	if rec.incs[0].route != rec.incs[1].route || rec.incs[1].route != rec.incs[2].route {
		t.Errorf("unknown routes produced different labels: %v (cardinality bomb)",
			[]string{rec.incs[0].route, rec.incs[1].route, rec.incs[2].route})
	}
}
