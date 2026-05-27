package o11y_test

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/srkn0/main/internal/o11y"
)

func TestLiveness_always200(t *testing.T) {
	h := o11y.LivenessHandler()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

func TestReadiness_200WhenReady(t *testing.T) {
	h := o11y.ReadinessHandler(func() error { return nil })
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

func TestReadiness_503WhenNotReady(t *testing.T) {
	h := o11y.ReadinessHandler(func() error { return errors.New("stores not loaded") })
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want 503", rec.Code)
	}
}

func TestReadiness_includesErrorInBody(t *testing.T) {
	// Body should at least contain the error so kubectl describe pod shows it.
	h := o11y.ReadinessHandler(func() error { return errors.New("stores not loaded") })
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))

	body, _ := io.ReadAll(rec.Body)
	if !strings.Contains(string(body), "stores not loaded") {
		t.Errorf("body = %q, want it to mention the error", body)
	}
}
