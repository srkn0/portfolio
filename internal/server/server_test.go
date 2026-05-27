package server

import (
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"syscall"
	"testing"
	"testing/fstest"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/srkn0/main/internal/o11y"
)

func testDeps() Deps {
	return Deps{
		Logger:   slog.New(slog.NewJSONHandler(io.Discard, nil)),
		Recorder: o11y.NewRecorder(prometheus.NewRegistry()),
	}
}

// fullDataFS builds a complete data FS with posts/, projects/ and cv/
// so loadStores succeeds end-to-end.
func fullDataFS() fs.FS {
	return fstest.MapFS{
		"posts/hello/de.md":   &fstest.MapFile{Data: []byte("---\ntitle: \"Hallo\"\ndate: 2026-05-01\n---\n")},
		"projects/demo/de.md": &fstest.MapFile{Data: []byte("---\ntitle: \"Demo\"\ndate: 2026-05-01\n---\n")},
		"cv/cv.de.md":         &fstest.MapFile{Data: []byte("# CV\n")},
	}
}

func publicFS() fs.FS {
	return fstest.MapFS{
		"css/index.css": &fstest.MapFile{Data: []byte("body{}")},
	}
}

func TestLoadStores_success(t *testing.T) {
	posts, projects, cv, err := loadStores(fullDataFS())
	if err != nil {
		t.Fatalf("loadStores: %v", err)
	}
	if posts == nil || projects == nil || cv == nil {
		t.Fatal("all three stores must be returned")
	}
}

func TestLoadStores_missingPostsDir(t *testing.T) {
	dataFS := fstest.MapFS{
		"projects/demo/de.md": &fstest.MapFile{Data: []byte("---\ntitle: \"Demo\"\ndate: 2026-05-01\n---\n")},
		"cv/cv.de.md":         &fstest.MapFile{Data: []byte("# CV\n")},
	}
	_, _, _, err := loadStores(dataFS)
	if err == nil {
		t.Fatal("expected error when posts/ is missing")
	}
}

func TestLoadStores_invalidPostFrontmatter(t *testing.T) {
	dataFS := fstest.MapFS{
		"posts/bad/de.md":     &fstest.MapFile{Data: []byte("---\ndate: not-a-date\n---\n")},
		"projects/demo/de.md": &fstest.MapFile{Data: []byte("---\ntitle: \"Demo\"\ndate: 2026-05-01\n---\n")},
		"cv/cv.de.md":         &fstest.MapFile{Data: []byte("# CV\n")},
	}
	_, _, _, err := loadStores(dataFS)
	if err == nil {
		t.Fatal("expected error for invalid post frontmatter")
	}
}

func TestPublicFileServer_servesAndSetsCacheControl(t *testing.T) {
	h := publicFileServer(publicFS())

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/public/css/index.css", nil))

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	if rec.Header().Get("Cache-Control") == "" {
		t.Error("expected Cache-Control header on static file")
	}
}

func TestNewRouter_rootRoutes(t *testing.T) {
	h := newTestHandlers(t)
	router := newRouter(h, publicFS(), testDeps(), func() error { return nil })

	for _, path := range []string{"/", "/blog", "/projects", "/cv", "/cv/print", "/contact"} {
		t.Run(path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
			if rec.Code != http.StatusOK {
				t.Errorf("%s -> %d, want 200", path, rec.Code)
			}
		})
	}
}

func TestNewRouter_unknownRouteIs404(t *testing.T) {
	h := newTestHandlers(t)
	router := newRouter(h, publicFS(), testDeps(), func() error { return nil })

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/nope-doesnt-exist", nil))

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestRunWithShutdown_returnsAfterSignal(t *testing.T) {
	srv := &http.Server{
		Addr: "127.0.0.1:0",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = io.WriteString(w, "ok")
		}),
	}

	done := make(chan error, 1)
	go func() {
		done <- runWithShutdown(srv, 500*time.Millisecond, slog.New(slog.NewJSONHandler(io.Discard, nil)))
	}()

	// Give the goroutines a moment to wire signal handling and Listen().
	time.Sleep(100 * time.Millisecond)

	// Trigger graceful shutdown by sending SIGTERM to ourselves.
	if err := syscall.Kill(syscall.Getpid(), syscall.SIGTERM); err != nil {
		t.Fatalf("could not send SIGTERM: %v", err)
	}

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("runWithShutdown returned %v, want nil", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("runWithShutdown did not return within 3s")
	}
}
