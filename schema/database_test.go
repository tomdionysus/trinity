package schema

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewDatabase(t *testing.T) {
	inst := NewDatabase("DBNAME")

	assert.NotNil(t, inst)
	assert.Equal(t, "DBNAME", inst.Name)
}
