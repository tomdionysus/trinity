package packets

import (
	ch "github.com/tomdionysus/consistenthash"
	"time"
)

const (
	CMD_HEARTBEAT    = 1
	CMD_DISTRIBUTION = 2
)

type Packet struct {
	Command   uint16
	ID        ch.NodeId
	RequestID [16]byte
	Sent      time.Time

	Payload interface{}
}

func NewPacket(command uint16, payload interface{}) *Packet {
	inst := &Packet{
		Command: command,
		ID:      ch.NewRandomNodeId(),
		Sent:    time.Now(),
		Payload: payload,
	}
	return inst
}

func NewResponsePacket(command uint16, requestid [16]byte, payload interface{}) *Packet {
	inst := &Packet{
		Command:   command,
		ID:        [16]byte(ch.NewRandomKey()),
		RequestID: requestid,
		Sent:      time.Now(),
		Payload:   payload,
	}
	return inst
}
