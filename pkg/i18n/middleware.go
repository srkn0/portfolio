package i18n

import (
	"net/http"
	"strings"
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
		Name: langCookie, Value: lang, Path: "/",
		MaxAge: 86400 * 365, HttpOnly: true, SameSite: http.SameSiteLaxMode,
	})
}

func detectLang(r *http.Request) string {
	if c, err := r.Cookie(langCookie); err == nil && IsValidLang(c.Value) {
		return c.Value
	}
	if qp := r.URL.Query().Get("lang"); IsValidLang(qp) {
		return qp
	}
	if accept := r.Header.Get("Accept-Language"); accept != "" {
		first := strings.TrimSpace(strings.Split(strings.Split(accept, ",")[0], ";")[0])
		primary := strings.ToLower(strings.Split(first, "-")[0])
		if IsValidLang(primary) {
			return primary
		}
	}
	return "de"
}
