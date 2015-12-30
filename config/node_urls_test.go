package config

import (
  "testing"
  "github.com/stretchr/testify/assert"
)

func TestNodeURLs(t *testing.T) {
  inst := &NodeURLs{}

  inst.Set("OK")
  assert.Equal(t, inst.String(), "[OK]")
  inst.Set("ANDAGAIN")
  assert.Equal(t, inst.String(), "[OK ANDAGAIN]")
}
