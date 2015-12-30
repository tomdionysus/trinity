package kvstore

import (
  "testing"
  "github.com/stretchr/testify/assert"
)

func TestNewKVStore(t *testing.T) { 
  inst := NewKVStore(nil)

  assert.NotNil(t, inst)
}