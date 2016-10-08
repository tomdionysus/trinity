package packets

import (
	"encoding/gob"
	"github.com/tomdionysus/consistenthash"
)

const (
	CMD_PEERLIST = 3
)

type PeerListPacket map[consistenthash.Key]string

func init() {
	gob.Register(PeerListPacket{})
}
