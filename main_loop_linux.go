package main

import(
  "os"
  "os/signal"
  "syscall"
  "github.com/tomdionysus/trinity/network"
  "github.com/tomdionysus/trinity/util"
)

// TrinityMainLoop for Linux
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
            logger.Info("Main", "Signal %d recieved, shutting down", sig)
            return
      }
    }
  }
}