package o11y

import (
	"fmt"
	"net/http"
)

// LivenessHandler answers "is the process up at all?". It does nothing
// besides return 200. No store checks, no DB pings — those belong in readiness.
//
// k8s liveness probe hits this. If it 5xx's, the pod gets killed.
func LivenessHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
}

// ReadinessFunc reports whether the app is ready to serve real traffic.
type ReadinessFunc func() error

// ReadinessHandler returns 200 when ready() returns nil, otherwise 503
// with the error message in the body.
//
// k8s readiness probe hits this. While it's 503, the pod is removed from
// the Service endpoints so it stops receiving traffic.
func ReadinessHandler(ready ReadinessFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		if err := ready(); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintf(w, "not ready: %v", err)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ready"))
	})
}
