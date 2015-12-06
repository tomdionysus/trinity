package network

import ( 
  "github.com/tomdionysus/trinity/util"
  "github.com/tomdionysus/trinity/packets"
  "crypto/tls"
  "errors"
  // "bytes"
  "encoding/gob"
)

const (
  PeerStateDisconnected = iota
  PeerStateConnecting = iota
  PeerStateHandshake = iota
  PeerStateConnected = iota
)

type Peer struct {
  Logger *util.Logger
  Server *TLSServer
  
  Address string
  State uint

  Connection *tls.Conn

  Writer *gob.Encoder
  Reader *gob.Decoder
}

func NewPeer(logger *util.Logger, server *TLSServer, address string) *Peer {
  inst := &Peer{
    Logger: logger,
    Address: address,
    State: PeerStateDisconnected,
    Server: server,
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

  go me.process()

  return nil
}

func (me *Peer) process() {

  var packet packets.Packet

  for {
    // Read Command
    err := me.Reader.Decode(&packet)
    if err!=nil {
      if err.Error()=="EOF" {
        // Disconnected.
        goto end
      }
      me.Logger.Error("Peer", "Error Reading: %s", err.Error())
      goto end
    }
    switch (packet.Command) {
      case packets.CMD_HEARTBEAT:
        me.Logger.Debug("Peer", "Got CMD_HEARTBEAT", packet.Command)
      case packets.CMD_DISTRIBUTION:
        me.Logger.Debug("Peer", "Got CMD_DISTRIBUTION: %s", packet.Payload)
      default:
        me.Logger.Warn("Peer", "Got Unknown Packet Command %d", packet.Command)
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