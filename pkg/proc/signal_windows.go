//go:build windows
// +build windows

package proc

import (
	"context"
	"os"
	"syscall"
)

var onlyOneSignalHandler = make(chan struct{})

func SetupSignalHandler() context.Context {
	close(onlyOneSignalHandler)
	c := make(chan os.Signal, 2)
	ctx, cancle := context.WithCancel(context.Background())
	signal.Notify(c, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)

	go func() {
		<-c
		cancle()
		<-c
		os.Exit(0)
	}()

	return ctx
}
