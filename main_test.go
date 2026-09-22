package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"
)

// freePort asks the kernel for an unused port and releases it. There is a small
// race before run() binds it, which is acceptable for a local test and far less
// flaky than hardcoding port numbers.
func freePort(t *testing.T) string {
	t.Helper()
	l, err := (&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserving a port: %v", err)
	}
	defer func() { _ = l.Close() }()
	return fmt.Sprintf("127.0.0.1:%d", l.Addr().(*net.TCPAddr).Port)
}

// TestRunServesAndShutsDownCleanly covers the lifecycle: both listeners come up,
// serve, and run() returns nil once the context is cancelled.
func TestRunServesAndShutsDownCleanly(t *testing.T) {
	addr, metricsAddr := freePort(t), freePort(t)
	t.Setenv("RANDOMON_ADDR", addr)
	t.Setenv("RANDOMON_METRICS_ADDR", metricsAddr)

	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() { done <- run(ctx) }()

	waitReady(t, "http://"+addr+"/healthz")
	waitReady(t, "http://"+metricsAddr+"/metrics")

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("run() = %v, want nil on a clean shutdown", err)
		}
	case <-time.After(shutdownTimeout + 5*time.Second):
		t.Fatal("run() did not return after the context was cancelled")
	}
}

// TestRunFailsOnBadImagePool proves the validation is wired into startup.
func TestRunFailsOnBadImagePool(t *testing.T) {
	t.Setenv("RANDOMON_ADDR", freePort(t))
	t.Setenv("RANDOMON_METRICS_ADDR", freePort(t))
	t.Setenv("RANDOMON_IMAGES", "/relative.gif")

	if err := run(t.Context()); err == nil {
		t.Error("run() = nil, want an error for a relative image URL")
	}
}

func TestEnvOr(t *testing.T) {
	if got := envOr("RANDOMON_TEST_UNSET", "fallback"); got != "fallback" {
		t.Errorf("envOr(unset) = %q, want %q", got, "fallback")
	}
	t.Setenv("RANDOMON_TEST_SET", "value")
	if got := envOr("RANDOMON_TEST_SET", "fallback"); got != "value" {
		t.Errorf("envOr(set) = %q, want %q", got, "value")
	}
}

func waitReady(t *testing.T, url string) {
	t.Helper()
	client := &http.Client{Timeout: time.Second}
	for range 100 {
		req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, url, nil)
		if err != nil {
			t.Fatalf("building request: %v", err)
		}
		resp, err := client.Do(req)
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return
			}
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("%s never became ready", url)
}
