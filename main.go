package main

import (
	// "github.com/tomdionysus/trinity/schema"
	"github.com/tomdionysus/trinity/server"
	// "github.com/tomdionysus/trinity/sql"
	"github.com/tomdionysus/trinity/util"
	"github.com/tomdionysus/trinity/config"
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

	// Banner
	logger.Info("Main", "---------------------------------------")
	logger.Info("Main", "Trinity DB - v%s", VERSION)
	logger.Info("Main", "---------------------------------------")

	// Config Debug
	logger.Debug("Config","Nodes: %s", config.Nodes.String())
	logger.Debug("Config","Certificate: %s", *config.Certificate)
	logger.Debug("Config","Port: %d", *config.Port)
	logger.Debug("Config","LogLevel: %s (%d)", *config.LogLevel, logger.LogLevel)

	// Server
	svr := server.NewTLSServer(logger)

	// Certificate
	err := svr.LoadPEMCert(*config.Certificate, *config.Certificate)
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

	c := make(chan os.Signal, 1)
  signal.Notify(c, os.Interrupt)
  signal.Notify(c, syscall.SIGTERM)

  for {
  	select {
  		case <-c:
  			logger.Info("Main", "SIGINT recieved, shutting down")
  			goto end
  	}
  }

  end:

  svr.Stop()
  _ = <-svr.StatusChannel

  logger.Info("Main", "Shutdown Complete, exiting")
  os.Exit(0)
}
