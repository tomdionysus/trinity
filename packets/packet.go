package packets

import (
  "time"
  "math/rand"
  // "encoding/gob"
)

const(
  PacketIDSize = 8

  CMD_HEARTBEAT = 1
  CMD_DISTRIBUTION = 2
)

type Packet struct {
  Command uint16
  ID [PacketIDSize]byte
  Sent time.Time

  Payload interface{}
}

func NewPacket(command uint16, payload interface{}) *Packet {
  inst := &Packet{
    Command: command,
    ID: GetRandomID(),
    Sent: time.Now(),
    Payload: payload,
  }
  return inst
}

func GetRandomID() [PacketIDSize]byte {
  rand.Seed(time.Now().UTC().UnixNano())
  b := [PacketIDSize]byte{}
  for i:=0; i<PacketIDSize; i++ {
    b[i] = byte(rand.Intn(256))
  }
  return b
}
