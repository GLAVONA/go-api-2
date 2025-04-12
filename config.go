package main

import (
	"time"

	"golang.org/x/time/rate"
)

type limiterConfig struct {
	rate  rate.Limit
	burst int
}

type config struct {
	usersCleanupTick time.Duration
	limiter          limiterConfig
	cleanupSince     time.Duration
}

var Config = &config{
	usersCleanupTick: 3 * time.Second,
	cleanupSince:     5 * time.Second,
	limiter: limiterConfig{
		rate:  1,
		burst: 3,
	},
}
