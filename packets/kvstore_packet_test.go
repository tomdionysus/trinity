package packets

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKVStorePacket(t *testing.T) {
	inst := &KVStorePacket{}

	assert.NotNil(t, inst)
}
