package kvstore

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tomdionysus/trinity/util"
)

func TestNewKVStore(t *testing.T) {
	inst := NewKVStore(nil)

	assert.NotNil(t, inst)
}

func TestInit(t *testing.T) {
	logger := util.NewLogger("error")
	inst := NewKVStore(logger)

	inst.Init()
}

func TestStartStop(t *testing.T) {
	logger := util.NewLogger("error")
	inst := NewKVStore(logger)
	inst.Start()
	assert.True(t, inst.running)
	inst.Stop()
	time.Sleep(1000 * time.Millisecond)
	assert.False(t, inst.running)
}

func TestGetSet(t *testing.T) {
	logger := util.NewLogger("error")
	inst := NewKVStore(logger)

	data := []byte{0, 1, 2}
	inst.Set("one", data, 1, nil)

	val, flag, ok := inst.Get("one")
	assert.Equal(t, val, data)
	assert.Equal(t, flag, int16(1))
	assert.True(t, ok)
	val, flag, ok = inst.Get("two")
	assert.Nil(t, val)
	assert.Equal(t, flag, int16(0))
	assert.False(t, ok)

}
