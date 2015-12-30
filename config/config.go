package config

import(
  "flag"
  "fmt"
  "errors"
)

type Config struct {
  Nodes NodeURLs
  CA *string
  Certificate *string
  Port *int
  LogLevel *string
  MemcacheEnabled *bool
  MemcachePort *int
  HostAddr *string
}

func NewConfig() *Config {
  inst := &Config{}

  flag.Var(&inst.Nodes, "node", "URL of another trinity node")
  inst.CA = flag.String("ca", "ca.pem", "CA PEM file")
  inst.Certificate = flag.String("cert", "cert.pem", "Certificate PEM file")
  inst.LogLevel = flag.String("loglevel", "error", "Logging Level [error,warn,info,debug]")
  inst.Port = flag.Int("port", 13531, "Cluster port")
  inst.MemcacheEnabled = flag.Bool("memcache", false, "Enable Memcache Server")
  inst.MemcachePort = flag.Int("memcacheport", 11211, "Memcache port")
  inst.HostAddr = flag.String("hostaddr", "localhost:13531", "Advertisted hostname:port")
  flag.Parse()
  return inst
}

func (me *Config) Validate() (bool, []error) {
  errs := []error{}
  if *me.Port<0 || *me.Port>65535 {
    errs = append(errs,errors.New(fmt.Sprintf("Port %d is invalid (0-65535)", *me.Port)))
  }
  return len(errs)==0, errs
}


