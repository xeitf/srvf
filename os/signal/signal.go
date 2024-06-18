package signal

import (
	"os"
	"os/signal"
)

// Wait
func Wait(signals ...os.Signal) {
	if len(signals) <= 0 {
		return
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, signals...)

	<-signalChan
}
