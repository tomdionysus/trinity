package packets

import (
  "time"
  // "encoding/gob"
  "github.com/tomdionysus/trinity/util"
)

const(
  CMD_HEARTBEAT = 1
  CMD_DISTRIBUTION = 2
  CMD_PEERLIST = 2

  PacketIDSize = 8
)

type Packet struct {
  Command uint16
  ID []byte
  Sent time.Time

  Payload interface{}
}

func NewPacket(command uint16, payload interface{}) *Packet {
  inst := &Packet{
    Command: command,
    ID: util.GetRandomID(PacketIDSize),
    Sent: time.Now(),
    Payload: payload,
  }
  return inst
}
