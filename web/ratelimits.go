package web

import (
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

var (
	rateLimitsMu sync.Mutex
	rateLimits   = make(map[string]*rateLimitInfos)
)

type rateLimitInfos struct {
	Remaining  uint64        // requests remaining
	ResetDelay time.Duration // time until limit reset
	LastUpdate time.Time
}

func (infos *rateLimitInfos) isExpired() bool {
	return !infos.LastUpdate.Add(infos.ResetDelay).After(time.Now())
}

// wait returns how long to wait before the next request: the time left
// until the limit resets, once at most one request remains.
func (infos *rateLimitInfos) wait() time.Duration {
	if infos.isExpired() || infos.Remaining > 1 {
		return 0
	}
	return time.Until(infos.LastUpdate.Add(infos.ResetDelay))
}

func updateRateLimits(response *http.Response) {
	if response.Header.Get("x-ratelimit-remaining") == "" || response.Header.Get("x-ratelimit-reset") == "" {
		return
	}

	limits := rateLimitInfos{
		LastUpdate: time.Now(),
	}

	// number of requests left for the time window
	remaining, err := strconv.ParseUint(response.Header.Get("x-ratelimit-remaining"), 10, 0)
	if err != nil {
		log.Println("Failed to parse ratelimit-remaining value", err, "sent by", response.Request.URL.Host)
		return
	}
	limits.Remaining = remaining

	// number of seconds before the rate limit resets
	reset, err := strconv.ParseUint(response.Header.Get("x-ratelimit-reset"), 10, 0)
	if err != nil {
		log.Println("Failed to parse ratelimit-reset value", err, "sent by", response.Request.URL.Host)
		return
	}
	limits.ResetDelay = time.Duration(reset) * time.Second

	// After a redirect, response.Request is the request to the target, whose
	// Host field is empty, so the key comes from its URL.
	rateLimitsMu.Lock()
	rateLimits[response.Request.URL.Host] = &limits
	rateLimitsMu.Unlock()
}

func awaitRateLimits(host string) {
	// Entries are replaced, never mutated, so the pointer is safe to read
	// after unlocking, and the sleep below does not block other hosts.
	rateLimitsMu.Lock()
	limits := rateLimits[host]
	rateLimitsMu.Unlock()

	if limits == nil {
		return
	}
	if d := limits.wait(); d > 0 {
		log.Printf("Rate limit of %s reached, waiting %s\n", host, d.Round(time.Second))
		time.Sleep(d)
	}
}
