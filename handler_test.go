package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

func newTestServer(images ...string) *server {
	if len(images) == 0 {
		images = []string{"https://example.test/a.gif"}
	}
	return &server{images: images}
}

func get(t *testing.T, h http.Handler, path string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, path, nil))
	return rec
}

func TestRandomImageRedirects(t *testing.T) {
	pool := []string{"https://example.test/a.gif", "https://example.test/b.gif"}
	rec := get(t, newTestServer(pool...).routes(), "/")

	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusFound)
	}
	if loc := rec.Header().Get("Location"); !slices.Contains(pool, loc) {
		t.Errorf("Location = %q, want one of %v", loc, pool)
	}
	if cc := rec.Header().Get("Cache-Control"); !strings.Contains(cc, "no-store") {
		t.Errorf("Cache-Control = %q, want it to contain no-store", cc)
	}
}

// TestRedirectsUseTheWholePool is the test that actually covers the product:
// asserting only membership would pass for a handler that always returned
// images[0], which would make randomon not random.
func TestRedirectsUseTheWholePool(t *testing.T) {
	pool := []string{"https://example.test/a.gif", "https://example.test/b.gif"}
	srv := newTestServer(pool...)
	h := srv.routes()

	seen := map[string]int{}
	for range 200 {
		seen[get(t, h, "/").Header().Get("Location")]++
	}
	for _, want := range pool {
		if seen[want] == 0 {
			t.Errorf("%q never served in 200 requests; seen = %v", want, seen)
		}
	}
}

func TestRedirectIncrementsCounter(t *testing.T) {
	before := testutil.ToFloat64(redirectsTotal)
	get(t, newTestServer().routes(), "/")

	if got := testutil.ToFloat64(redirectsTotal) - before; got != 1 {
		t.Errorf("randomon_redirects_total increased by %v, want 1", got)
	}
}

func TestHealthz(t *testing.T) {
	rec := get(t, newTestServer().routes(), "/healthz")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

// TestMetricsNotOnPublicMux guards the split introduced after the security
// review: the gateway publishes the public listener with PathPrefix "/", so
// /metrics there would expose the Go runtime collector (exact Go version,
// restart time, memory) to the internet. Importing client_golang also pulls in
// expvar, which registers /debug/vars on http.DefaultServeMux.
func TestMetricsNotOnPublicMux(t *testing.T) {
	h := newTestServer().routes()
	for _, path := range []string{"/metrics", "/debug/vars"} {
		rec := get(t, h, path)
		if rec.Code != http.StatusFound {
			t.Errorf("%s: status = %d, want %d (should be a plain redirect)", path, rec.Code, http.StatusFound)
		}
		if body := rec.Body.String(); strings.Contains(body, "go_info") || strings.Contains(body, "memstats") {
			t.Errorf("%s: public mux leaked runtime data", path)
		}
	}
}

func TestMetricsServedOnItsOwnMux(t *testing.T) {
	rec := get(t, metricsRoutes(), "/metrics")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), "randomon_redirects_total") {
		t.Error("metrics output is missing randomon_redirects_total")
	}
}

func TestDefaultImagesAreAbsoluteAndNotSelfReferential(t *testing.T) {
	for _, raw := range defaultImages {
		u, err := url.Parse(raw)
		if err != nil {
			t.Errorf("image %q does not parse: %v", raw, err)
			continue
		}
		if u.Scheme != "https" {
			t.Errorf("image %q: scheme = %q, want https", raw, u.Scheme)
		}
		if u.Host == "randomon.crstian.me" {
			t.Errorf("image %q points back at randomon itself", raw)
		}
	}
}
