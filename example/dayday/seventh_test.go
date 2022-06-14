package dayday

import (
	"context"
	"golang.org/x/time/rate"
	"sort"
	"time"
)

type RateLimiter interface {
	Wait(ctx context.Context) error
	Limit() rate.Limit
}

func MultiLimiter(limiters ...RateLimiter) *multiLimiter {
	byLimit := func(i, j int) bool {
		return limiters[i].Limit() < limiters[j].Limit()
	}
	sort.Slice(limiters, byLimit)
	return &multiLimiter{limiters: limiters}
}

type multiLimiter struct {
	limiters []RateLimiter
}

func (l *multiLimiter) Wait(ctx context.Context) error {
	//TODO implement me
	for _, l := range l.limiters {
		if err := l.Wait(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (l *multiLimiter) Limit() rate.Limit {
	//TODO implement me
	return l.limiters[0].Limit()
}

func NewAPIConnectionV2() *APIConnectionV2 {
	secondLimit := rate.NewLimiter(Per(2, time.Second), 1)
	minuteLimit := rate.NewLimiter(Per(10, time.Minute), 10)
	return &APIConnectionV2{
		rateLimiter:  MultiLimiter(secondLimit, minuteLimit),
		diskLimit:    MultiLimiter(rate.NewLimiter(rate.Limit(1), 1)),
		networkLimit: MultiLimiter(rate.NewLimiter(Per(3, time.Second), 3)),
	}
}

type APIConnectionV2 struct {
	networkLimit,
	diskLimit,
	rateLimiter RateLimiter
}

func (a *APIConnectionV2) ReadFile(ctx context.Context) error {
	if err := MultiLimiter(a.rateLimiter, a.diskLimit).Wait(ctx); err != nil {
		return err
	}
	// Pretend we do work here
	return nil
}

func (a *APIConnectionV2) ResolveAddress(ctx context.Context) error {
	if err := MultiLimiter(a.rateLimiter, a.networkLimit).Wait(ctx); err != nil {
		return err
	}
	// Pretend we do work here
	return nil
}

func Per(eventCount int, duration time.Duration) rate.Limit {
	return rate.Every(duration / time.Duration(eventCount))
}
