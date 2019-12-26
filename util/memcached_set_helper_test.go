package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemcachedSetArgsHelperCommonUse(t *testing.T) {
	args := make([]string, 6)
	args[0] = "set"
	args[1] = "test"
	args[2] = "0"
	args[3] = "100"
	args[4] = "0"

	expire, flags, bytes, err := MemcachedSetArgsHelper(args)
	assert.Nil(t, err)
	assert.Equal(t, expire, 100)
	assert.Equal(t, flags, 0)
	assert.Equal(t, bytes, 0)
}

func TestMemcachedSetArgsHelperIntToBig(t *testing.T) {
	args := make([]string, 6)
	args[0] = "set"
	args[1] = "test"
	args[2] = "0"
	args[3] = "10000000000000000000000000000"
	args[4] = "0"

	expire, flags, bytes, err := MemcachedSetArgsHelper(args)
	assert.NotNil(t, err)
	assert.Equal(t, expire, 0)
	assert.Equal(t, flags, 0)
	assert.Equal(t, bytes, 0)
}

func TestMemcachedSetArgsHelperErrorNotNumber(t *testing.T) {
	args := make([]string, 6)
	args[0] = "set"
	args[1] = "test"
	args[2] = "AAAAAAAAAAAAAAAAA"
	args[3] = "100"
	args[4] = "0"

	expire, flags, bytes, err := MemcachedSetArgsHelper(args)
	assert.NotNil(t, err)
	assert.Equal(t, expire, 0)
	assert.Equal(t, flags, 0)
	assert.Equal(t, bytes, 0)
}
