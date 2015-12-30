package util

import (
  "testing"
  "github.com/stretchr/testify/assert"
)

func TestNewLogger(t *testing.T) {
  inst := NewLogger("error")

  assert.NotNil(t, inst)
  assert.Equal(t, uint(4), inst.LogLevel)
} 