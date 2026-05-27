package i18n_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/srkn0/main/pkg/i18n"
)

func TestIsValidLang(t *testing.T) {
	cases := map[string]bool{
		"de": true,
		"en": true,
		"fr": false,
		"":   false,
		"DE": false, // strict
	}
	for in, want := range cases {
		if got := i18n.IsValidLang(in); got != want {
			t.Errorf("IsValidLang(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestSetLangCookie(t *testing.T) {
	rec := httptest.NewRecorder()
	i18n.SetLangCookie(rec, "en")

	res := rec.Result()
	cookies := res.Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie, got %d", len(cookies))
	}
	c := cookies[0]
	if c.Name != "lang" {
		t.Errorf("Name = %q, want lang", c.Name)
	}
	if c.Value != "en" {
		t.Errorf("Value = %q, want en", c.Value)
	}
	if c.Path != "/" {
		t.Errorf("Path = %q, want /", c.Path)
	}
	if !c.HttpOnly {
		t.Error("HttpOnly should be true")
	}
	if c.SameSite != http.SameSiteLaxMode {
		t.Errorf("SameSite = %v, want Lax", c.SameSite)
	}
	if c.MaxAge <= 0 {
		t.Errorf("MaxAge = %d, want > 0", c.MaxAge)
	}
}

func TestMiddleware_injectsLocaleFromCookie(t *testing.T) {
	captured := ""
	next := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		captured = i18n.GetLocale(r.Context())
	})
	handler := i18n.Middleware(next)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "lang", Value: "en"})
	handler.ServeHTTP(httptest.NewRecorder(), req)

	if captured != "en" {
		t.Errorf("locale = %q, want en", captured)
	}
}

func TestMiddleware_injectsLocaleFromQueryParam(t *testing.T) {
	captured := ""
	next := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		captured = i18n.GetLocale(r.Context())
	})
	handler := i18n.Middleware(next)

	req := httptest.NewRequest(http.MethodGet, "/?lang=en", nil)
	handler.ServeHTTP(httptest.NewRecorder(), req)

	if captured != "en" {
		t.Errorf("locale = %q, want en", captured)
	}
}

func TestMiddleware_injectsLocaleFromAcceptLanguage(t *testing.T) {
	cases := []struct {
		header string
		want   string
	}{
		{"en-US,en;q=0.9", "en"},
		{"de-DE,de;q=0.9,en;q=0.8", "de"},
		{"fr,en;q=0.8", i18n.DefaultLocale}, // fr not supported, primary kicks back
	}
	for _, tc := range cases {
		t.Run(tc.header, func(t *testing.T) {
			captured := ""
			next := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
				captured = i18n.GetLocale(r.Context())
			})
			handler := i18n.Middleware(next)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Accept-Language", tc.header)
			handler.ServeHTTP(httptest.NewRecorder(), req)

			if captured != tc.want {
				t.Errorf("locale = %q, want %q", captured, tc.want)
			}
		})
	}
}

func TestMiddleware_fallbackToDefault(t *testing.T) {
	captured := ""
	next := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		captured = i18n.GetLocale(r.Context())
	})
	handler := i18n.Middleware(next)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(httptest.NewRecorder(), req)

	if captured != i18n.DefaultLocale {
		t.Errorf("locale = %q, want %q", captured, i18n.DefaultLocale)
	}
}

func TestMiddleware_cookieBeatsQueryAndHeader(t *testing.T) {
	captured := ""
	next := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		captured = i18n.GetLocale(r.Context())
	})
	handler := i18n.Middleware(next)

	req := httptest.NewRequest(http.MethodGet, "/?lang=en", nil)
	req.AddCookie(&http.Cookie{Name: "lang", Value: "de"})
	req.Header.Set("Accept-Language", "en")
	handler.ServeHTTP(httptest.NewRecorder(), req)

	if captured != "de" {
		t.Errorf("locale = %q, want de (cookie wins)", captured)
	}
}

func TestMiddleware_invalidCookieFallsThrough(t *testing.T) {
	captured := ""
	next := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		captured = i18n.GetLocale(r.Context())
	})
	handler := i18n.Middleware(next)

	req := httptest.NewRequest(http.MethodGet, "/?lang=en", nil)
	req.AddCookie(&http.Cookie{Name: "lang", Value: "fr"}) // invalid
	handler.ServeHTTP(httptest.NewRecorder(), req)

	if captured != "en" {
		t.Errorf("locale = %q, want en (fall back to query)", captured)
	}
}
