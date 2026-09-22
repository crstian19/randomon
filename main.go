// Command randomon is a tiny HTTP service that answers every request with a
// redirect to a randomly chosen image URL.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	defaultAddr        = ":3000"
	defaultMetricsAddr = ":9090"
	readHeaderTimeout  = 5 * time.Second
	readTimeout        = 10 * time.Second
	writeTimeout       = 15 * time.Second
	idleTimeout        = 60 * time.Second
	// Must exceed writeTimeout: a request that runs to the write deadline would
	// otherwise consume the whole shutdown budget. Still well under the 30s
	// terminationGracePeriod Kubernetes gives us.
	shutdownTimeout = 20 * time.Second
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel()})))

	if err := run(context.Background()); err != nil {
		slog.Error("randomon exited with an error", "error", err)
		os.Exit(1)
	}
}

func logLevel() slog.Level {
	var lvl slog.Level
	if err := lvl.UnmarshalText([]byte(strings.ToUpper(envOr("LOG_LEVEL", "INFO")))); err != nil {
		return slog.LevelInfo
	}
	return lvl
}

func run(ctx context.Context) error {
	// SIGTERM is what Kubernetes sends before killing the pod.
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	images, err := loadImages()
	if err != nil {
		return fmt.Errorf("load images: %w", err)
	}

	// Two listeners on purpose. The gateway route publishes the public one with
	// PathPrefix "/", so anything served there is reachable from the internet;
	// metrics stay on a separate port that is only scraped from inside the
	// cluster.
	public := newServer(envOr("RANDOMON_ADDR", defaultAddr), (&server{images: images}).routes())
	metrics := newServer(envOr("RANDOMON_METRICS_ADDR", defaultMetricsAddr), metricsRoutes())

	errCh := make(chan error, 2)
	serve := func(name string, srv *http.Server) {
		slog.Info("listening", "server", name, "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("%s listener: %w", name, err)
		}
	}
	go serve("public", public)
	go serve("metrics", metrics)

	var runErr error
	select {
	case runErr = <-errCh:
	case <-ctx.Done():
		slog.Info("shutdown signal received")
	}
	// Restore the default signal handling so a second SIGTERM kills the process
	// instead of being swallowed while we drain.
	stop()

	// WithoutCancel: the signal already cancelled ctx, but shutdown still needs
	// its own live deadline to drain in-flight requests.
	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), shutdownTimeout)
	defer cancel()

	// Concurrently: sharing one deadline sequentially lets a slow public
	// request starve the metrics shutdown and turn a clean stop into exit 1.
	var wg sync.WaitGroup
	errs := make([]error, 2)
	for i, srv := range []*http.Server{public, metrics} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs[i] = srv.Shutdown(shutdownCtx)
		}()
	}
	wg.Wait()

	// Drain a listener error that raced with the shutdown signal.
	select {
	case err := <-errCh:
		runErr = errors.Join(runErr, err)
	default:
	}

	if err := errors.Join(errs[0], errs[1], runErr); err != nil {
		return err
	}
	slog.Info("randomon stopped cleanly")
	return nil
}

func newServer(addr string, h http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           h,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
