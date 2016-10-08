package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNodeURLs(t *testing.T) {
	inst := &NodeURLs{}

	inst.Set("OK")
	assert.Equal(t, inst.String(), "[OK]")
	inst.Set("ANDAGAIN")
	assert.Equal(t, inst.String(), "[OK ANDAGAIN]")
}
