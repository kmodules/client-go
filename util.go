package kutil

import "time"

const (
	RetryInterval = 50 * time.Millisecond
	RetryTimeout  = 2 * time.Second
)
