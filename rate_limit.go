package parse

import "time"

type limiter interface {
	limit()
}

type rateLimiterT struct {
	c chan time.Time
}

func newRateLimiter(limit, burst uint) *rateLimiterT {
	r := &rateLimiterT{
		c: make(chan time.Time, burst),
	}
	go func() {
		for t := range time.Tick(time.Second / time.Duration(limit)) {
			r.c <- t
		}
	}()

	return r
}

func (l *rateLimiterT) limit() {
	<-l.c
}
