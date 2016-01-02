package main

import (
	// "github.com/tomdionysus/trinity/schema"
	// "github.com/tomdionysus/trinity/sql"
  "github.com/tomdionysus/trinity/network"
	"github.com/tomdionysus/trinity/util"
  "github.com/tomdionysus/trinity/config"
  "github.com/tomdionysus/trinity/kvstore"
	// "github.com/tomdionysus/trinity/packets"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	// Load Configuration
	config := config.NewConfig()

	// Start logging
	logger := util.NewLogger(*config.LogLevel)

	// Validate Configuration
	if i, errs := config.Validate(); !i {
		for _, err := range errs {
			logger.Error("Config", err.Error())
		}
		os.Exit(-1)
		logger.Fatal("Main", "Bad Configuration, Exiting")
	}

  // Key/Value Store
  kv := kvstore.NewKVStore(logger)
  kv.Init()
  kv.Start()

	// Banner
	logger.Raw("Main", "---------------------------------------")
	logger.Raw("Main", "Trinity DB - v%s", VERSION)
	logger.Raw("Main", "---------------------------------------")

	// Config Debug
	logger.Debug("Config","Nodes: %s", config.Nodes.String())
	logger.Debug("Config","Certificate: %s", *config.Certificate)
	logger.Debug("Config","Port: %d", *config.Port)
	logger.Debug("Config","LogLevel: %s (%d)", *config.LogLevel, logger.LogLevel)

  // CA
  capool := network.NewCAPool(logger)
  err := capool.LoadPEM(*config.CA)
  if err != nil {
    logger.Error("Main", "Cannot Load CA '%s': %s", *config.CA, err.Error())
    os.Exit(-1)
  }
  logger.Debug("Main", "CA Certiticate Loaded") 

	// Server
	svr := network.NewTLSServer(logger, capool, kv, *config.HostAddr)
  logger.Info("Main","Trinity Node ID %02X", svr.ServerNode.ID)

	// Certificate
	err = svr.LoadPEMCert(*config.Certificate, *config.Certificate)
	if err != nil {
		logger.Error("Main", "Cannot Load Certificate '%s': %s", *config.Certificate, err.Error())
		os.Exit(-1)
	}
	logger.Debug("Main", "Cert Loaded")	

	// Listen
	err = svr.Listen(uint16(*config.Port))
	if err != nil {
		logger.Error("Main", "Cannot Start Server: %s", err.Error())
		os.Exit(-1)
	}

  var memcache *network.MemcacheServer

  // Memcache
  if *config.MemcacheEnabled {
    memcache = network.NewMemcacheServer(logger, *config.MemcachePort, svr)
    memcache.Init()
    memcache.Start()
  }

	// Notify SIGINT, SIGTERM
	c := make(chan os.Signal, 1)
  signal.Notify(c, os.Interrupt)
  signal.Notify(c, syscall.SIGTERM)
  // signal.Notify(c, syscall.SIGINFO) // syscall.SIGINFO doesn't exist in linux go.

  // TEST: Connect to other nodes
  for _, remoteAddr := range config.Nodes {
    svr.ConnectTo(remoteAddr)
  }

	// Wait for SIGINT
  for {
  	select {
  		case sig := <-c:
        switch sig {
          // case syscall.SIGINFO:
          //   logger.Info("Main", "Status: Node ID %02X", svr.ServerNode.ID)
          //   logger.Info("Main", "Status: Listener Address %s", svr.Listener.Addr())
          //   for _, peer := range svr.Connections { 
          //     logger.Info("Main", "Status: Peer %02X (%s -> %s), %s", peer.ServerNetworkNode.ID, peer.Connection.LocalAddr(), peer.Connection.RemoteAddr(), network.PeerStateString[peer.State])
          //   }
          case os.Interrupt:
            fallthrough
          case syscall.SIGTERM:
            logger.Info("Main", "Signal %d recieved, shutting down", sig)
            goto end
      }
  	}
  }

  end:

  // Shutdown Memcache
  if memcache!=nil { memcache.Stop() }

  // Shutdown Server and wait for close
  svr.Stop()
  _ = <-svr.StatusChannel

  // Shutdown KV Store
  kv.Stop()

  logger.Info("Main", "Shutdown Complete, exiting")
  os.Exit(0)
}
