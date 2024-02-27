package util

import (
	"time"

	"github.com/cenkalti/backoff/v4"
)

func NewBackoff(timeout time.Duration) backoff.BackOff {
	b := backoff.NewExponentialBackOff()
	b.InitialInterval = 500 * time.Millisecond
	b.MaxInterval = 5 * time.Second
	b.MaxElapsedTime = timeout
	return b
}
