package crawl_limiter

import (
	"context"
	"golang.org/x/time/rate"
	"net/url"
	"sync"
	"time"
)

type CrawlLimiter struct {
	mu       sync.Mutex
	maxReqs  int
	window   time.Duration
	limiters map[string]*rate.Limiter
}

func NewCrawlLimiter(maxReqs int, window time.Duration) *CrawlLimiter {
	return &CrawlLimiter{
		maxReqs:  maxReqs,
		window:   window,
		limiters: make(map[string]*rate.Limiter),
	}
}

func (c *CrawlLimiter) Wait(site *url.URL, ctx context.Context) {
	c.mu.Lock()

	host := site.Hostname()

	l, ok := c.limiters[host]

	if !ok {
		l = rate.NewLimiter(rate.Every(c.window), c.maxReqs)
		c.limiters[host] = l
	}

	c.mu.Unlock()

	l.Wait(ctx)
}
