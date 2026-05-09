package i18n

import (
	"net/http"
	"strings"
	"time"
)

const langCookie = "lang"

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := WithLocale(r.Context(), detectLang(r))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func IsValidLang(lang string) bool {
	return lang == "de" || lang == "en"
}

func SetLangCookie(w http.ResponseWriter, lang string) {
	http.SetCookie(w, &http.Cookie{
		Name:     langCookie,
		Value:    lang,
		Path:     "/",
		MaxAge:   int((365 * 24 * time.Hour).Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func detectLang(r *http.Request) string {
	if c, err := r.Cookie(langCookie); err == nil && IsValidLang(c.Value) {
		return c.Value
	}

	if qp := r.URL.Query().Get("lang"); IsValidLang(qp) {
		return qp
	}

	if lang := primaryAcceptLanguage(r.Header.Get("Accept-Language")); IsValidLang(lang) {
		return lang
	}

	return DefaultLocale
}

func primaryAcceptLanguage(header string) string {
	if header == "" {
		return ""
	}
	first := strings.Split(header, ",")[0]
	first = strings.TrimSpace(strings.Split(first, ";")[0])
	return strings.ToLower(strings.Split(first, "-")[0])
}
