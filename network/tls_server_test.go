package network

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTLSServer(t *testing.T) {
	inst := NewTLSServer(nil, nil, nil, "HOSTNAME")

	assert.NotNil(t, inst)
}
