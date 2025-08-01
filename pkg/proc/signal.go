//go:build linux || darwin
// +build linux darwin

package proc

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/mangohow/cloud-ide/pkg/logger"
)

var onlyOneSignalHandler = make(chan struct{})

func SetupSignalHandler() context.Context {
	close(onlyOneSignalHandler) // panics when called twice

	ctx, cancel := context.WithCancel(context.Background())

	c := make(chan os.Signal, 2)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGUSR1, syscall.SIGUSR2)

	setLogger(logger.Logger())
	var stopper Stopper

	go func() {
		for i := 0; i < 2; {
			select {
			case sig := <-c:
				switch sig {
				case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
					if i == 0 {
						cancel()
					} else {
						os.Exit(0)
					}
					i++
				case syscall.SIGUSR1:
					if stopper == nil {
						stopper = StartProfile()
					} else {
						stopper.Stop()
						stopper = nil
					}
				case syscall.SIGUSR2:
					dumpGoroutines()
				}
			}
		}
	}()

	return ctx
}
