package main

import (
	"log/slog"
	"math/rand/v2"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var redirectsTotal = promauto.NewCounter(prometheus.CounterOpts{
	Name: "randomon_redirects_total",
	Help: "Total number of image redirects served.",
})

// server holds the image pool on the struct so tests can exercise routes()
// against a fixture pool instead of the compiled-in one.
type server struct {
	images []string
}

// routes builds the PUBLIC handler. It must always return its own ServeMux:
// importing client_golang pulls in expvar, whose init() registers /debug/vars
// on http.DefaultServeMux. Since the gateway route publishes this listener with
// PathPrefix "/", returning DefaultServeMux (or passing a nil Handler to
// http.Server) would put /debug/vars on the public internet.
func (s *server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.handleHealth)
	// Catch-all: any other path answers with a redirect, which is what the
	// <img src="https://randomon.crstian.me/"> use case relies on.
	mux.HandleFunc("/", s.handleRandomImage)
	return mux
}

// metricsRoutes builds the handler for the internal metrics listener, kept off
// the public port so the Go runtime collector (exact Go version, restart time,
// memory) is not readable from the internet.
func metricsRoutes() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /metrics", promhttp.Handler())
	return mux
}

func (s *server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok\n"))
}

// handleRandomImage answers with a 302 to a randomly chosen image. Caching is
// disabled so an <img> tag reloaded by a browser or proxied by a CDN gets a
// fresh pick every time.
func (s *server) handleRandomImage(w http.ResponseWriter, r *http.Request) {
	//nolint:gosec // G404: picking a GIF at random is not a security decision.
	target := s.images[rand.IntN(len(s.images))]

	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	redirectsTotal.Inc()
	slog.Debug("redirecting", "target", target, "path", r.URL.Path)

	http.Redirect(w, r, target, http.StatusFound)
}
