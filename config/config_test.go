package config

import (
  "testing"
  "github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
  inst := NewConfig()

  assert.Equal(t, "ca.pem", *inst.CA)
  assert.Equal(t, "cert.pem", *inst.Certificate)
  assert.Equal(t, "error", *inst.LogLevel)
  assert.Equal(t, 13531, *inst.Port)
  assert.Equal(t, false, *inst.MemcacheEnabled)
  assert.Equal(t, 11211, *inst.MemcachePort)
  assert.Equal(t, "localhost:13531", *inst.HostAddr)

  // Defaults should validate OK
  ok, errs := inst.Validate()
  assert.Equal(t, []error{}, errs)
  assert.True(t, ok)

  // But if the port is out of band..
  *inst.Port = 65538
  ok, errs = inst.Validate()
  assert.NotNil(t, errs)
  assert.False(t, ok)
}