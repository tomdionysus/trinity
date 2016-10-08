package network

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTLSServer(t *testing.T) {
	inst := NewTLSServer(nil, nil, nil, "HOSTNAME")

	assert.NotNil(t, inst)
}
