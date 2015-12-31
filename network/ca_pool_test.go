package network

import (
  "testing"
  "github.com/stretchr/testify/assert"
  "github.com/tomdionysus/trinity/util"
)

func TestNewCAPool(t *testing.T) {
  inst := NewCAPool(nil)

  assert.NotNil(t, inst)
  assert.NotNil(t, inst.Pool)
  assert.Nil(t, inst.Logger)
} 

func TestLoadPEMNoFile(t *testing.T) {
  inst := NewCAPool(util.NewLogger("fatal"))

  err := inst.LoadPEM("iugvWROCBWqufbhOIGWfc9qgiucwb")

  assert.NotNil(t, err)
} 