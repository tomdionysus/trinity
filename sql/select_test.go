package sql

import (
  "testing"
  "github.com/stretchr/testify/assert"
)

func TestSelect(t *testing.T) {
  inst := &Select{}

  assert.NotNil(t, inst)
}