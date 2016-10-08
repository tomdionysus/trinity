package packets

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestKVStorePacket(t *testing.T) {
	inst := &KVStorePacket{}

	assert.NotNil(t, inst)
}
