package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestClient_ConcurrentRateLimitedRequests(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("x-ratelimit-remaining", "300")
		w.Header().Set("x-ratelimit-reset", "60")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c := NewClient()
	var wg sync.WaitGroup
	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var v map[string]any
			if err := c.GetJson(srv.URL, &v, nil); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
}

// A limit that resets in 60 s, updated 50 s ago, has 10 s left.
func TestRateLimitWaitsOnlyTheTimeLeft(t *testing.T) {
	limits := rateLimitInfos{Remaining: 1, ResetDelay: 60 * time.Second, LastUpdate: time.Now().Add(-50 * time.Second)}

	got := limits.wait()

	if got <= 0 || got > 10*time.Second {
		t.Errorf("wait() = %s, want at most 10s", got)
	}
}

func TestRateLimitWaitsNotWithRequestsLeft(t *testing.T) {
	limits := rateLimitInfos{Remaining: 2, ResetDelay: 60 * time.Second, LastUpdate: time.Now()}

	if got := limits.wait(); got != 0 {
		t.Errorf("wait() = %s, want 0", got)
	}
}

// The limit of a redirect target is recorded under that host, which the
// next request to it looks up.
func TestRateLimitAfterRedirectKeysTheTargetHost(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("x-ratelimit-remaining", "0")
		w.Header().Set("x-ratelimit-reset", "60")
		_, _ = w.Write([]byte(`{}`))
	}))
	defer target.Close()
	origin := httptest.NewServer(http.RedirectHandler(target.URL, http.StatusFound))
	defer origin.Close()

	var v map[string]any
	if err := NewClient().GetJson(origin.URL, &v, nil); err != nil {
		t.Fatal(err)
	}

	host := strings.TrimPrefix(target.URL, "http://")
	rateLimitsMu.Lock()
	limits, ok := rateLimits[host]
	_, empty := rateLimits[""]
	rateLimitsMu.Unlock()
	if !ok || limits.Remaining != 0 {
		t.Errorf("rateLimits[%s] = %+v, %v, want the target's limit", host, limits, ok)
	}
	if empty {
		t.Error("a limit was recorded under the empty host")
	}
}
