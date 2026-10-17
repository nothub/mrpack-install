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
		log.Println("Failed to parse ratelimit-remaining value", err, "sent by", response.Request.Host)
		return
	}
	limits.Remaining = remaining

	// number of seconds before the rate limit resets
	reset, err := strconv.ParseUint(response.Header.Get("x-ratelimit-reset"), 10, 0)
	if err != nil {
		log.Println("Failed to parse ratelimit-reset value", err, "sent by", response.Request.Host)
		return
	}
	limits.ResetDelay = time.Duration(reset) * time.Second

	rateLimitsMu.Lock()
	rateLimits[response.Request.Host] = &limits
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
	if limits.isExpired() {
		return
	}
	if limits.Remaining <= 1 {
		log.Println("Awaiting rate limits for:", host, "Sleeping for", limits.ResetDelay, "seconds...")
		time.Sleep(limits.ResetDelay)
	}
}
