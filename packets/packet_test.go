package packets

import (
  "testing"
  "github.com/stretchr/testify/assert"
  "github.com/tomdionysus/consistenthash"
)

func TestNewPacket(t *testing.T) {
  payload := "Test"
  inst := NewPacket(CMD_PEERLIST,payload)

  assert.Equal(t, inst.Command, uint16(CMD_PEERLIST))
  assert.NotNil(t, inst.ID)
  assert.NotNil(t, inst.Sent)
  assert.Equal(t, inst.Payload, payload)
}

func TestNewResponsePacket(t *testing.T) {
  payload := "Test2"

  k1 := [16]byte(consistenthash.NewRandomKey())
  inst := NewResponsePacket(CMD_DISTRIBUTION, k1, payload)

  assert.Equal(t, inst.Command, uint16(CMD_DISTRIBUTION))
  assert.NotNil(t, inst.ID)
  assert.NotNil(t, inst.RequestID, k1)
  assert.NotNil(t, inst.Sent)
  assert.Equal(t, inst.Payload, payload)
}