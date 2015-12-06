package network

import ( 
  "github.com/tomdionysus/trinity/util"
  "crypto/tls"
  "errors"
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

func (me *Peer) Start() {
  if me.State != PeerStateHandshake {
    me.Logger.Error("Peer", "Cannot Start Client, Handshake not ready")
    return
  }
  err := me.Connection.Handshake()
  if err!=nil {
    me.Logger.Error("Peer", "Peer TLS Handshake failed, disconnecting: %s",err.Error())
    me.Disconnect()
    return
  }
  state := me.Connection.ConnectionState()
  if len(state.PeerCertificates)==0 {
    me.Logger.Error("Peer", "Peer has no certificates, disconnecting")
    me.Disconnect()
    return
  }
  sub := state.PeerCertificates[0].Subject.CommonName
  me.Logger.Info("Peer", "Connected to %s (%s) [%s]", me.Connection.RemoteAddr(), sub, Ciphers[me.Connection.ConnectionState().CipherSuite])
  me.State = PeerStateConnected
}