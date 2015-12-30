package schema

import (
  "testing"
  "github.com/stretchr/testify/assert"
)

func TestNewDatabase(t *testing.T) {
  inst := NewDatabase("DBNAME")

  assert.NotNil(t, inst)
  assert.Equal(t, "DBNAME", inst.Name)
}