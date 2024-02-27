package util

import (
	"os"
	"os/signal"
	"syscall"
)

func NewInterruptChan() <-chan os.Signal {
	var stopChan = make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	return stopChan
}
