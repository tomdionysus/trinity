package network

import (
  "testing"
  "github.com/stretchr/testify/assert"
)

func TestNewCAPool(t *testing.T) {
  inst := NewCAPool(nil)

  assert.NotNil(t, inst)
  assert.NotNil(t, inst.Pool)
  assert.Nil(t, inst.Logger)
} 