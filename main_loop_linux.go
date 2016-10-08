package main

import (
	"github.com/tomdionysus/trinity/network"
	"github.com/tomdionysus/trinity/util"
	"os"
	"os/signal"
	"syscall"
)

// TrinityMainLoop for Linux (No SIGINFO)
func TrinityMainLoop(svr *network.TLSServer, logger *util.Logger) {
	// Notify SIGINT, SIGTERM
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)

	// Wait for SIGINT
	for {
		select {
		case sig := <-c:
			switch sig {
			case os.Interrupt:
				fallthrough
			case syscall.SIGTERM:
				logger.Info("Main", "Signal %d received, shutting down", sig)
				return
			}
		}
	}
}
