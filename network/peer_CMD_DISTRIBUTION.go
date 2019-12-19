package network

import (
	"github.com/tomdionysus/consistenthash"
	"github.com/tomdionysus/trinity/packets"
	"time"
)

// process_CMD_DISTRIBUTION processes a CMD_DISTRIBUTION packet received from a peer.
// The packet contains the ID and the CH distribution of the peer. It is the last stage of the
// connection establishment protocol, and a peer is not PeerStateConnected until this packet is received and verified.
func (peer *Peer) process_CMD_DISTRIBUTION(packet packets.Packet) {
	// CMD_DISTRIBUTION should only be received once, right at the start, for both incoming and outgoing
	// connections.
	if peer.ServerNetworkNode != nil {
		peer.Logger.Warn("Peer", "%02X: CMD_DISTRIBUTION received from registered peer", peer.ServerNetworkNode.ID)
		return
	}
	node := packet.Payload.(consistenthash.ServerNetworkNode)
	peer.ServerNetworkNode = &node
	peer.Server.ConnectionSet(peer.ServerNetworkNode.ID, peer)
	peer.Logger.Debug("Peer", "%02X: CMD_DISTRIBUTION (%s)", peer.ServerNetworkNode.ID, peer.Connection.RemoteAddr())

	// Because of misconfiguration, complex network topologies, or another faulty peer, it's possible that
	// this connection is actually from ourselves. If so, shut it down.
	if peer.Server.ServerNode.ID.EqualTo(peer.ServerNetworkNode.ID) {
		if !peer.Incoming {
			peer.Logger.Warn("Peer", "%02X: The outgoing connection connected back to this node - disconnecting (%s)", peer.ServerNetworkNode.ID, peer.Connection.RemoteAddr())
		} else {
			peer.Logger.Debug("Peer", "%02X: Incoming connection is from this node - disconnecting (%s)", peer.ServerNetworkNode.ID, peer.Connection.RemoteAddr())
		}
		peer.Disconnect()
		return
	}

	// The connection may be from a node we're already connected to, If so, shut it down.
	if peer.Server.ServerNode.NodeRegistered(peer.ServerNetworkNode.ID) {
		// Peer has previously registered
		peer.Logger.Debug("Peer", "%02X: Node Already Registered", peer.ServerNetworkNode.ID)
	} else {
		// Peer is new
		err := peer.Server.ServerNode.RegisterNode(peer.ServerNetworkNode)
		if err != nil {
			peer.Logger.Error("Peer", "%02X: Register Node Distribution Failed: %s", peer.ServerNetworkNode.ID, err.Error())
			return
		}
	}

	// Peer is now connected, update the heartbeat
	peer.State = PeerStateConnected
	peer.LastHeartbeat = time.Now()

	// And ask the server to notify all peers of all other peers
	peer.Server.NotifyAllPeers()
}
