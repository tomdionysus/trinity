package network

import (
	"github.com/tomdionysus/trinity/packets"
)

// process_CMD_PEERLIST processes a CMD_PEERLIST packet received from a peer.
// The packet contains a list of all remote instances connected to that peer, with the
// exception of the receiving and sending peers for this packet.
func (peer *Peer) process_CMD_PEERLIST(packet packets.Packet) {
	peers := packet.Payload.(packets.PeerListPacket)
	peer.Logger.Debug("Peer", "%02X: CMD_PEERLIST (%d Peers)", peer.ServerNetworkNode.ID, len(peers))

	for id, k := range peers {
		if peer.Server.ServerNode.ID.EqualTo(id) {
			peer.Logger.Warn("Peer", "%02X: - Notified us of ourselves (%02X).", peer.ServerNetworkNode.ID, id)
		} else {
			if _, connected := peer.Server.ConnectionGet(id); !connected {
				peer.Logger.Debug("Peer", "%02X: - Notified %02X - Connecting New Peer (%s)", peer.ServerNetworkNode.ID, id, k)
				peer.Server.ConnectTo(k)
			} else {
				peer.Logger.Debug("Peer", "%02X: - Notified %02X - Already Connected to Peer (%s)", peer.ServerNetworkNode.ID, id, k)
			}
		}
	}
}
