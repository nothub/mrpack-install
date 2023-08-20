package http

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
)

var rateLimits = make(map[string]*rateLimitInfos)

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
		fmt.Println("Failed to parse ratelimit-remaining value", err, "sent by", response.Request.Host)
		return
	}
	limits.Remaining = remaining

	// number of seconds before the rate limit resets
	reset, err := strconv.ParseUint(response.Header.Get("x-ratelimit-reset"), 10, 0)
	if err != nil {
		fmt.Println("Failed to parse ratelimit-reset value", err, "sent by", response.Request.Host)
		return
	}
	limits.ResetDelay = time.Duration(reset) * time.Second

	rateLimits[response.Request.Host] = &limits
}

func awaitRateLimits(host string) {
	if rateLimits[host] == nil {
		return
	}
	if rateLimits[host].isExpired() {
		return
	}
	if rateLimits[host].Remaining <= 1 {
		fmt.Println("Awaiting rate limits for:", host, "Sleeping for", rateLimits[host].ResetDelay, "seconds...")
		time.Sleep(rateLimits[host].ResetDelay)
	}
}
