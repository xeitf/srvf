package process

import (
	"os"
	"os/signal"
)

// WaitSignal
func WaitSignal(signals ...os.Signal) {
	if len(signals) <= 0 {
		return
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, signals...)

	for range signalChan {
		return
	}
}
