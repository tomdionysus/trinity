package packets

import (
	ch "github.com/tomdionysus/consistenthash"
	"time"
)

const (
	CMD_HEARTBEAT    = 1
	CMD_DISTRIBUTION = 2
)

type PacketId ch.Key

func NewRandomPacketId() PacketId {
	return PacketId(ch.NewRandomKey())
}

type Packet struct {
	Command   uint16
	ID        PacketId
	RequestID PacketId
	Sent      time.Time

	Payload interface{}
}

func NewPacket(command uint16, payload interface{}) *Packet {
	inst := &Packet{
		Command: command,
		ID:      NewRandomPacketId(),
		Sent:    time.Now(),
		Payload: payload,
	}
	return inst
}

func NewResponsePacket(command uint16, requestid PacketId, payload interface{}) *Packet {
	inst := &Packet{
		Command:   command,
		ID:        NewRandomPacketId(),
		RequestID: requestid,
		Sent:      time.Now(),
		Payload:   payload,
	}
	return inst
}
