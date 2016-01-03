package main

import(
  "os"
  "os/signal"
  "syscall"
  "github.com/tomdionysus/trinity/network"
  "github.com/tomdionysus/trinity/util"
)

// TrinityMainLoop for Darwin (MacOSX) includes SIGINFO (Ctrl-T) signal for status.
func TrinityMainLoop(svr *network.TLSServer, logger *util.Logger) {
  // Notify SIGINT, SIGTERM
  c := make(chan os.Signal, 1)
  signal.Notify(c, os.Interrupt)
  signal.Notify(c, syscall.SIGTERM)
  signal.Notify(c, syscall.SIGINFO) // syscall.SIGINFO doesn"t exist in linux go.

  iostatus := map[bool]string{ true: "Incoming", false: "Outgoing" }
  // Wait for SIGINT
  for {
    select {
      case sig := <-c:
        switch sig {
          case syscall.SIGINFO:
            logger.Info("Main", "Status: Node ID %02X", svr.ServerNode.ID)
            logger.Info("Main", "Status: Listener Address %s", svr.Listener.Addr())
            for _, peer := range svr.Connections { 
              logger.Info("Main", "Status: Peer %02X (%s %s) %s", peer.ServerNetworkNode.ID, iostatus[peer.Incoming], peer.Connection.RemoteAddr(), network.PeerStateString[peer.State])
            }
          case os.Interrupt:
            fallthrough
          case syscall.SIGTERM:
            logger.Info("Main", "Signal %d recieved, shutting down", sig)
            return
      }
    }
  }
}