package shutdown

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/pchchv/sws/helpers/ancli"
)

// Monitor listens for a shutdown signal and cancels the context if a signal is received.
// If the signal is received again, it will force a shutdown.
// It is aborted when ctx is canceled.
func Monitor(ctx context.Context, cancel context.CancelFunc) {
	var amountOfCancels int
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case <-ctx.Done():
			return
		case <-signalCh:
			if amountOfCancels == 0 {
				ancli.PrintWarn("initiating shutdown")
				cancel()
			} else if amountOfCancels == 1 {
				ancli.PrintWarn("graceful shutdown ongoing, cancel again to force shutdown")
			} else {
				ancli.PrintErr("forcing shutdown")
				os.Exit(1)
			}
			amountOfCancels++
		}
	}
}
