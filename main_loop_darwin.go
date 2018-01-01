package main

import (
	"github.com/tomdionysus/trinity/network"
	"github.com/tomdionysus/trinity/util"
	"os"
	"os/signal"
	"syscall"
)

// TrinityMainLoop for Darwin (MacOSX) includes SIGINFO (Ctrl-T) signal for status.
func TrinityMainLoop(svr *network.TLSServer, logger *util.Logger) {
	// Notify SIGINT, SIGTERM
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	signal.Notify(c, syscall.SIGINFO) // syscall.SIGINFO doesn't exist in linux go.

	logger.Info("Main", "MacOSX - Use (Ctrl-T) for status")

	iostatus := map[bool]string{true: "Incoming", false: "Outgoing"}
	// Wait for SIGINT
	for {
		select {
		case sig := <-c:
			switch sig {
			case syscall.SIGINFO:
				logger.Info("Main", "Status: Node ID %02X", svr.ServerNode.ID)
				logger.Info("Main", "Status: Listener Address %s", svr.Listener.Addr())
				logger.Info("Main", "Status: Advertised Address %s", svr.ServerNode.HostAddr)
				connections := svr.Connections()
				logger.Info("Main", "Status: %d Active Connection(s)", len(connections))
				for _, peer := range connections {
					logger.Info("Main", "Status: Peer %02X (%s %s) %s", peer.ServerNetworkNode.ID, iostatus[peer.Incoming], peer.Connection.RemoteAddr(), network.PeerStateString[peer.State])
				}
			case os.Interrupt:
				fallthrough
			case syscall.SIGTERM:
				logger.Info("Main", "Signal %d received, shutting down", sig)
				return
			}
		}
	}
}
