package packets

import (
	"time"

	ch "github.com/tomdionysus/consistenthash"
)

const (
	CMD_HEARTBEAT    = 1
	CMD_DISTRIBUTION = 2
	CMD_STATUS_SYNC  = 3
)

// PacketId is an alias for storing a consistenthash key
type PacketId ch.Key

// NewRandomPacketId Generate a new id for sending packet
func NewRandomPacketId() PacketId {
	return PacketId(ch.NewRandomKey())
}

// Packet represent a system packet
type Packet struct {
	Command   uint16
	ID        PacketId
	RequestID PacketId
	Sent      time.Time

	Payload interface{}
}

// NewPacket Generate a packet which are not going to wait for a response
func NewPacket(command uint16, payload interface{}) *Packet {
	inst := &Packet{
		Command: command,
		ID:      NewRandomPacketId(),
		Sent:    time.Now(),
		Payload: payload,
	}
	return inst
}

// NewResponsePacket Generate a packet which are going to wait for a response
func NewResponsePacket(command uint16, requestID PacketId, payload interface{}) *Packet {
	inst := &Packet{
		Command:   command,
		ID:        NewRandomPacketId(),
		RequestID: requestID,
		Sent:      time.Now(),
		Payload:   payload,
	}
	return inst
}
