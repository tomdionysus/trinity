package packets

import (
  "time"
)

const(
  CMD_KVSTORE = 10

  CMD_KVSTORE_SET = 1 
  CMD_KVSTORE_GET = 1 
  CMD_KVSTORE_DELETE = 1 
)

type KVStorePacket struct {
  KVCommand int16

  Key string
  Data []byte
  ExpiresAt time.Time
  Flags int16

  TargetID [16]byte
}
