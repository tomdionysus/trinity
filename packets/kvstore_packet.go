package packets

import (
	"encoding/gob"
	"time"

	ch "github.com/tomdionysus/consistenthash"
)

const (
	CMD_KVSTORE           = 10
	CMD_KVSTORE_ACK       = 11
	CMD_KVSTORE_NOT_FOUND = 12

	CMD_KVSTORE_SET    = 1
	CMD_KVSTORE_GET    = 2
	CMD_KVSTORE_DELETE = 3
	CMD_KVSTORE_IS_SET = 4
)

type KVStorePacket struct {
	Command int16

	Key       string
	KeyHash   [16]byte
	Data      []byte
	ExpiresAt *time.Time
	Flags     int16

	TargetID ch.NodeId
}

func init() {
	gob.Register(KVStorePacket{})
}
