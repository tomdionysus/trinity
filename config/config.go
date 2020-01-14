package config

import (
	"flag"
	"fmt"
)

// Config struct hold config information for the node
type Config struct {
	Nodes            NodeURLs
	CA               *string
	Certificate      *string
	Port             *int
	LogLevel         *string
	MemcacheEnabled  *bool
	MemcachePort     *int
	HostAddr         *string
	DisableHeartbeat *bool
}

// NewConfig init a new Config struct with default value
func NewConfig() *Config {
	inst := &Config{}

	flag.Var(&inst.Nodes, "node", "URL of another trinity node")
	inst.CA = flag.String("ca", "ca.pem", "CA PEM file")
	inst.Certificate = flag.String("cert", "cert.pem", "Certificate PEM file")
	inst.LogLevel = flag.String("loglevel", "error", "Logging Level [error,warn,info,debug]")
	inst.Port = flag.Int("port", 13531, "Cluster port")
	inst.MemcacheEnabled = flag.Bool("memcache", false, "Enable Memcache Server")
	inst.MemcachePort = flag.Int("memcacheport", 11211, "Memcache port")
	inst.HostAddr = flag.String("hostaddr", "", "Advertised hostname:port")
	inst.DisableHeartbeat = flag.Bool("disable-heartbeat", false, "[DEV ONLY] Disable heartbeat check to avoid losing connection on breakpoint")
	flag.Parse()

	if *inst.HostAddr == "" {
		s := fmt.Sprintf("localhost:%d", *inst.Port)
		inst.HostAddr = &s
	}

	return inst
}

// Validate configuration
func (cfg *Config) Validate() (bool, []error) {
	errs := []error{}
	if *cfg.Port < 0 || *cfg.Port > 65535 {
		errs = append(errs, fmt.Errorf("Port %d is invalid (0-65535)", *cfg.Port))
	}
	return len(errs) == 0, errs
}
