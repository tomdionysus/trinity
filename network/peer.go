package network

import ( 
  "github.com/tomdionysus/trinity/util"
  "github.com/tomdionysus/trinity/packets"
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
  PeerStateDefib = iota
)

type Peer struct {
  Logger *util.Logger
  Server *TLSServer
  
  Address string
  State uint

  Connection *tls.Conn
  HeartbeatTicker *time.Ticker

  Writer *gob.Encoder
  Reader *gob.Decoder

  LastHeartbeat time.Time
}

func NewPeer(logger *util.Logger, server *TLSServer, address string) *Peer {
  inst := &Peer{
    Logger: logger,
    Address: address,
    State: PeerStateDisconnected,
    Server: server,
    LastHeartbeat: time.Now(),
  }
  return inst
}

func NewConnectingPeer(logger *util.Logger, server *TLSServer, connection *tls.Conn) *Peer {
  inst := NewPeer(logger, server, connection.RemoteAddr().String())
  inst.Connection = connection
  inst.State = PeerStateHandshake
  return inst
}

func (me *Peer) Connect() error {
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
  me.State = PeerStateDisconnected
  me.Connection.Close()
  me.HeartbeatTicker.Stop()
  me.Logger.Info("Peer", "Disconnected: %s", me.Connection.RemoteAddr())
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
  me.State = PeerStateConnected

  me.Reader = gob.NewDecoder(me.Connection)
  me.Writer = gob.NewEncoder(me.Connection)

  go me.heartbeat()
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
      me.Logger.Warn("Peer", "%s: Peer in Defib (no response for 5 seconds)", me.Connection.RemoteAddr())
      me.State = PeerStateDefib
    }

    switch me.State {
      case PeerStateConnected:
        err := me.SendPacket(packets.NewPacket(packets.CMD_HEARTBEAT,nil))
        if err!=nil {
          me.Logger.Error("Peer","Error Sending Heartbeat, disconnecting", me.Connection.RemoteAddr())
          me.Disconnect()
          return
        }
      case PeerStateDefib:
        if time.Now().After(me.LastHeartbeat.Add(10 * time.Second)) {
          me.Logger.Warn("Peer", "%s: Peer DOA (Defib for 5 seconds, disconnecting)", me.Connection.RemoteAddr())
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
      me.Logger.Error("Peer", "%s: Error Reading: %s", me.Connection.RemoteAddr(), err.Error())
      goto end
    }
    switch (packet.Command) {
      case packets.CMD_HEARTBEAT:
        // me.Logger.Debug("Peer", "%s: CMD_HEARTBEAT", me.Connection.RemoteAddr())
        me.LastHeartbeat = time.Now()
      case packets.CMD_DISTRIBUTION:
        me.Logger.Debug("Peer", "%s: CMD_DISTRIBUTION: %s", me.Connection.RemoteAddr(), packet.Payload)
      default:
        me.Logger.Warn("Peer", "%s: Unknown Packet Command %d", me.Connection.RemoteAddr(), packet.Command)
    }
  }
  end:

  me.Disconnect()
}

func (me *Peer) SendPacket(packet *packets.Packet) error {
  err := me.Writer.Encode(packet)
  if err!=nil {
    me.Logger.Error("Peer", "Error Writing: %s", err.Error())
  }
  return err
}