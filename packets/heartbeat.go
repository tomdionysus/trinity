package packets

import (
  "time"
)

const(
  CMD_HEARTBEAT = 1
)

type Heartbeat BasePacket

func NewHeartbeat() *Heartbeat {
  inst := &Heartbeat{
    ID: GetRandomID()
    Command: CMD_HEARTBEAT
    Sent: time.Now()
  }
}

