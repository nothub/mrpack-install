package requester

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
	Timestamp  time.Time
}

func (infos *rateLimitInfos) IsExpired() bool {
	return !infos.Timestamp.Add(infos.ResetDelay).After(time.Now())
}

func updateRateLimits(response *http.Response) {
	if response.Header.Get("x-ratelimit-remaining") == "" ||
		response.Header.Get("x-ratelimit-reset") == "" {
		return
	}

	limits := rateLimitInfos{
		Timestamp: time.Now(),
	}

	remaining, err := strconv.ParseUint(response.Header.Get("x-ratelimit-remaining"), 10, 0)
	if err != nil {
		fmt.Println("Failed to parse ratelimit-remaining value", err, "sent by", response.Request.Host)
		return
	}
	limits.Remaining = remaining

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
	if rateLimits[host].IsExpired() {
		return
	}
	if rateLimits[host].Remaining <= 1 {
		fmt.Println("Awaiting rate limits for:", host, "Sleeping for", rateLimits[host].ResetDelay, "seconds...")
		time.Sleep(rateLimits[host].ResetDelay)
	}
}
