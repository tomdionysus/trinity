package network

import ( 
  "github.com/tomdionysus/trinity/util"
  "github.com/tomdionysus/trinity/packets"
  "github.com/tomdionysus/consistenthash"
  "crypto/tls"
  "errors"
  // "bytes"
  "encoding/gob"
  "time"
)

const (
  PeerStateDisconnected = iota
  PeerStateConnecting = iota
  PeerStateHandshake = iota
  PeerStateConnected = iota
  PeerStateSyncing = iota
  PeerStateDefib = iota
)

var PeerStateString map[uint]string = map[uint]string{
  PeerStateDisconnected: "PeerStateDisconnected",
  PeerStateConnecting: "PeerStateConnecting",
  PeerStateHandshake: "PeerStateHandshake",
  PeerStateConnected: "PeerStateConnected",
  PeerStateSyncing: "PeerStateSyncing",
  PeerStateDefib: "PeerStateDefib",
}

type Peer struct {
  Logger *util.Logger
  Server *TLSServer

  Incoming bool
  
  Address string
  State uint

  Connection *tls.Conn
  HeartbeatTicker *time.Ticker

  Writer *gob.Encoder
  Reader *gob.Decoder

  LastHeartbeat time.Time

  ServerNetworkNode *consistenthash.ServerNetworkNode

  Replies map[[16]byte]chan(*packets.Packet)
}

func NewPeer(logger *util.Logger, server *TLSServer, address string) *Peer {
  inst := &Peer{
    Logger: logger,
    Address: address,
    State: PeerStateDisconnected,
    Server: server,
    LastHeartbeat: time.Now(),
    ServerNetworkNode: nil,
    Replies: map[[16]byte]chan(*packets.Packet){},
  }
  return inst
}

func NewConnectingPeer(logger *util.Logger, server *TLSServer, connection *tls.Conn) *Peer {
  inst := NewPeer(logger, server, connection.RemoteAddr().String())
  inst.Connection = connection
  inst.State = PeerStateHandshake
  inst.Incoming = true
  return inst
}

func (me *Peer) Connect() error {
  me.Incoming = false
  me.State = PeerStateConnecting
  conn, err := tls.Dial("tcp", me.Address, &tls.Config{
    RootCAs: me.Server.CAPool.Pool,
    Certificates: []tls.Certificate{*me.Server.Certificate},
  })
  if err != nil {
    me.Logger.Error("Peer", "Cannot connect to %s: %s", me.Address, err.Error())
    me.Disconnect()
    return err
  }
  me.Connection = conn
  state := conn.ConnectionState()
  if len(state.PeerCertificates)==0 {
    me.Logger.Error("Peer", "Cannot connect to %s: Peer has no certificates", me.Address)
    me.Disconnect()
    return errors.New("Peer has no certificates")
  }
  me.State = PeerStateHandshake
  return nil
}


func (me *Peer) Disconnect() {
  if me.State == PeerStateConnected {
    me.State = PeerStateDisconnected
    me.Server.ServerNode.DeregisterNode(me.ServerNetworkNode)
    if me.HeartbeatTicker!=nil { me.HeartbeatTicker.Stop() }
    if me.Connection!=nil { me.Connection.Close() }
    delete(me.Server.Connections, me.ServerNetworkNode.ID)
    me.Logger.Info("Peer", "%02X: Disconnected", me.ServerNetworkNode.ID)
  }
}

func (me *Peer) Start() error {
  if me.State != PeerStateHandshake {
    me.Logger.Error("Peer", "Cannot Start Client, Handshake not ready")
    return errors.New("Handshake not ready")
  }
  err := me.Connection.Handshake()
  if err!=nil {
    me.Logger.Error("Peer", "Peer TLS Handshake failed, disconnecting: %s",err.Error())
    me.Disconnect()
    return errors.New("TLS Handshake Failed")
  }
  state := me.Connection.ConnectionState()
  if len(state.PeerCertificates)==0 {
    me.Logger.Error("Peer", "Peer has no certificates, disconnecting")
    me.Disconnect()
    return errors.New("Peer sent no certificates")
  }
  sub := state.PeerCertificates[0].Subject.CommonName
  me.Logger.Info("Peer", "Connected to %s (%s) [%s]", me.Connection.RemoteAddr(), sub, Ciphers[me.Connection.ConnectionState().CipherSuite])

  me.Reader = gob.NewDecoder(me.Connection)
  me.Writer = gob.NewEncoder(me.Connection)

  go me.heartbeat()

  me.SendDistribution()

  go me.process()

  return nil
}

// Ping the Peer every second.
func (me *Peer) heartbeat() {
  me.HeartbeatTicker = time.NewTicker(time.Second)

  for {
    <- me.HeartbeatTicker.C

    // Check For Defib
    if time.Now().After(me.LastHeartbeat.Add(5 * time.Second)) {
      me.Logger.Warn("Peer", "%02X: Peer Defib (no response for >5 seconds)", me.ServerNetworkNode.ID)
      me.State = PeerStateDefib
    }

    switch me.State {
      case PeerStateConnected:
        err := me.SendPacket(packets.NewPacket(packets.CMD_HEARTBEAT,nil))
        if err!=nil {
          me.Logger.Error("Peer", "%02X: Error Sending Heartbeat, disconnecting", me.ServerNetworkNode.ID)
          me.Disconnect()
          return
        }
      case PeerStateDefib:
        if time.Now().After(me.LastHeartbeat.Add(10 * time.Second)) {
          me.Logger.Warn("Peer", "%02X: Peer DOA (Defib for 5 seconds, disconnecting)", me.ServerNetworkNode.ID)
          me.Disconnect()
          return
        }
    }

  }

}

func (me *Peer) process() {

  var packet packets.Packet

  for {
    // Read Command
    err := me.Reader.Decode(&packet)
    if err!=nil {
      if err.Error()=="EOF" {
        me.Logger.Debug("Peer", "%02X: Peer Closed Connection", me.ServerNetworkNode.ID)
      } else {
        me.Logger.Error("Peer", "%02X: Error Reading: %s", me.ServerNetworkNode.ID, err.Error())
      }
      goto end
    }
    switch (packet.Command) {

      // Packets in Connecting / Handshake

      case packets.CMD_HEARTBEAT:
        me.LastHeartbeat = time.Now()
      case packets.CMD_DISTRIBUTION:
        if me.ServerNetworkNode!=nil {
          me.Logger.Warn("Peer", "%02X: CMD_DISTRIBUTION received from registered peer", me.ServerNetworkNode.ID)
          break
        } 
        servernetworknode := packet.Payload.(consistenthash.ServerNetworkNode)
        me.ServerNetworkNode = &servernetworknode
        me.Server.Connections[me.ServerNetworkNode.ID] = me
        me.Logger.Debug("Peer", "%02X: CMD_DISTRIBUTION (%s)", me.ServerNetworkNode.ID, me.Connection.RemoteAddr())
        redistribution, err := me.Server.ServerNode.RegisterNode(me.ServerNetworkNode)
        me.State = PeerStateConnected
        if err!=nil {
          me.Logger.Error("Peer","%02X: Register Node Distribution Failed: %s", me.ServerNetworkNode.ID, err.Error())
          break
        }
        if len(redistribution) > 0 {
          // for _, redist := range redistribution {
          //   me.Logger.Debug("Peer","Redistribution: %02x - %02x :  %02X -> %02X", redist.Start, redist.End, redist.SourceNodeID, redist.DestinationNodeID)
          // }
        }
        me.Server.NotifyNewPeer(me)
      // Packets in Connected

      case packets.CMD_KVSTORE:
        me.Logger.Debug("Peer", "%02X: CMD_KVSTORE", me.ServerNetworkNode.ID)
        me.handleKVStorePacket(&packet)
      case packets.CMD_KVSTORE_ACK:
        me.Logger.Debug("Peer", "%02X: CMD_KVSTORE_ACK", me.ServerNetworkNode.ID)
        me.handleReply(&packet)
      case packets.CMD_KVSTORE_NOT_FOUND:
        me.Logger.Debug("Peer", "%02X: CMD_KVSTORE_NOT_FOUND", me.ServerNetworkNode.ID)
        me.handleReply(&packet)
      case packets.CMD_PEERLIST:
        peers := packet.Payload.(packets.PeerListPacket)
        me.Logger.Debug("Peer", "%02X: CMD_PEERLIST (%d Peers)", me.ServerNetworkNode.ID, len(peers))
        for id, k := range peers {
          if me.Server.Listener.Addr().String() == k {
            me.Logger.Warn("Peer", "%02X: - Peer %02X (%s) notified us of ourselves.", id, me.ServerNetworkNode.ID, k)       
          } else {
            if !me.Server.IsConnectedTo(id)  {
              me.Logger.Debug("Peer", "%02X: - Connecting New Peer %02X (%s)", id, me.ServerNetworkNode.ID, k)
              me.Server.ConnectTo(k)
            } else {
              me.Logger.Debug("Peer", "%02X: - Already Connected to Peer %02X (%s)", id, me.ServerNetworkNode.ID, k)
            }
          }
        }
      default:
        me.Logger.Warn("Peer", "%02X: Unknown Packet Command %d", me.ServerNetworkNode.ID, packet.Command)
    }
  }
  end:

  me.Disconnect()
}

func (me *Peer) handleReply(packet *packets.Packet) {
  chn, found := me.Replies[packet.RequestID]
  if found {
    delete(me.Replies, packet.RequestID)
    chn <- packet
  } else {
    me.Logger.Warn("Peer", "%02X: Unsolicited Reply to unknown packet %02X", me.ServerNetworkNode.ID, packet.RequestID)
  }
}

func (me *Peer) SendDistribution() error {
  packet := packets.NewPacket(packets.CMD_DISTRIBUTION, me.Server.ServerNode.ServerNetworkNode)
  me.SendPacket(packet)
  return nil
}

func (me *Peer) SendPacket(packet *packets.Packet) error {
  err := me.Writer.Encode(packet)
  if err!=nil {
    me.Logger.Error("Peer", "Error Writing: %s", err.Error())
  }
  return err
}

func (me *Peer) SendPacketWaitReply(packet *packets.Packet, timeout time.Duration) (*packets.Packet, error) {
  if me.State != PeerStateConnected {
    me.Logger.Error("Peer", "%02X: Cannot send packet ID %02X, not PeerStateConnected", me.ServerNetworkNode.ID, packet.ID)
    return nil, errors.New("Cannot send, state not PeerStateConnected")
  }

  me.Replies[packet.ID] = make(chan(*packets.Packet))
  me.SendPacket(packet)
  reply := <- me.Replies[packet.ID]
  me.Logger.Debug("Peer", "%02X: Got Reply %02X for packet ID %02X", me.ServerNetworkNode.ID, reply.ID, packet.ID)
  return reply, nil
}

func (me *Peer) handleKVStorePacket(packet *packets.Packet) {
  kvpacket := packet.Payload.(packets.KVStorePacket)
  switch kvpacket.Command {
  case packets.CMD_KVSTORE_SET:
    me.handleKVStoreSet(&kvpacket, packet)
  case packets.CMD_KVSTORE_GET:
    me.handleKVStoreGet(&kvpacket, packet)
  case packets.CMD_KVSTORE_DELETE:
    me.handleKVStoreDelete(&kvpacket, packet)
  default:
    me.Logger.Error("Peer", "KVStorePacket: Unknown Command %d", packet.Command)
  }
}

func (me *Peer) handleKVStoreSet(packet *packets.KVStorePacket, request *packets.Packet) {
  me.Logger.Debug("Peer", "%02X: KVStoreSet: %s = %s", me.ServerNetworkNode.ID, packet.Key, packet.Data)
  me.Server.KVStore.Set(
    packet.Key,
    packet.Data,
    packet.Flags,
    packet.ExpiresAt)

  response := packets.NewResponsePacket(packets.CMD_KVSTORE_ACK, request.ID, packet.Key)
    me.Logger.Debug("Peer", "%02X: KVStoreSet: %s Acknowledge, replying", me.ServerNetworkNode.ID, packet.Key)
  me.SendPacket(response)
}

func (me *Peer) handleKVStoreGet(packet *packets.KVStorePacket, request *packets.Packet) {
  me.Logger.Debug("Peer", "%02X: KVStoreGet: %s", me.ServerNetworkNode.ID, packet.Key)
  value, flags, found := me.Server.KVStore.Get(packet.Key)

  var response *packets.Packet

  if found {
    payload := packets.KVStorePacket{
      Command: packets.CMD_KVSTORE_GET,
      Key: packet.Key,
      Data: value,
      Flags: flags,
    }
    response = packets.NewResponsePacket(packets.CMD_KVSTORE_ACK, request.ID, payload)
    me.Logger.Debug("Peer", "%02X: KVStoreGet: %s = %s, replying", me.ServerNetworkNode.ID, packet.Key, value)
  } else {
    response = packets.NewResponsePacket(packets.CMD_KVSTORE_NOT_FOUND, request.ID, packet.Key)
    me.Logger.Debug("Peer", "%02X: KVStoreGet: %s Not found, replying", me.ServerNetworkNode.ID, packet.Key)
  }

  me.SendPacket(response)
}

func (me *Peer) handleKVStoreDelete(packet *packets.KVStorePacket, request *packets.Packet) {
  me.Logger.Debug("Peer", "%02X: KVStoreDelete: %s", me.ServerNetworkNode.ID, packet.Key)
  found := me.Server.KVStore.Delete(packet.Key)
  
  var response *packets.Packet

  if found {
    response = packets.NewResponsePacket(packets.CMD_KVSTORE_ACK, request.ID, packet.Key)
    me.Logger.Debug("Peer", "%02X: KVStoreDelete: %s Deleted, replying", me.ServerNetworkNode.ID, packet.Key)
  } else {
    response = packets.NewResponsePacket(packets.CMD_KVSTORE_NOT_FOUND, request.ID, packet.Key)
    me.Logger.Debug("Peer", "%02X: KVStoreDelete: %s Not found, replying", me.ServerNetworkNode.ID, packet.Key)
  }

  me.SendPacket(response)
}





