package packets

import (
  "time"
  "math/rand"
)

const(
  PacketIDSize = 8
)

type Packet interface {
  GetCommand() uint16
  GetID() [PacketIDSize]byte
  GetSent() time.Time
}

type BasePacket struct {
  Command uint16
  ID [PacketIDSize]byte
  Sent time.Time
}

func (me *BasePacket) GetCommand() uint16 {
  return me.Command
}

func (me *BasePacket) GetID() [PacketIDSize]byte {
  return me.ID
}

func (me *BasePacket) GetSent() time.Time {
  return me.Sent
}

func GetRandomID() [PacketIDSize]byte {
  rand.Seed(time.Now().UTC().UnixNano())
  b := []byte{}
  for i:=0; i<size; i++ {
    b = append(b, byte(rand.Intn(256)))
  }
  x := bt.ByteSliceKey(b[:])
  return &x
}
