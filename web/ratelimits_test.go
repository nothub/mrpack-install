package web

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
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
