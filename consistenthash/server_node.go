package consistenthash

import(
  "math/rand"
  "time"
  bt "github.com/tomdionysus/binarytree"
  "errors"
)

const(
  DISTRIBUTION_MAX = 512
  NETWORK_ID_SIZE_BYTES = 16
  KEY_SIZE_BYTES = 16
)

type NodeDistribution [DISTRIBUTION_MAX]bt.ByteSliceKey

type ServerNetworkNode struct {
  ID bt.ByteSliceKey
  HostAddr string
  Distribution NodeDistribution
}

type ServerNode struct {
  ServerNetworkNode
  NetworkNodes map[bt.ByteSliceKey]*ServerNetworkNode
  Network *bt.Tree

  Values map[bt.ByteSliceKey]string
}

func NewServerNode(hostAddr string) *ServerNode {
  node := &ServerNode{ 
    Values: map[bt.ByteSliceKey]string{},
    NetworkNodes: map[bt.ByteSliceKey]*ServerNetworkNode{}, 
    Network: bt.NewTree(),
  }
  node.ID = RandKey()
  node.HostAddr = hostAddr
  node.Init()
  return node
}

func (me *ServerNode) Init() {
  me.Distribution = NodeDistribution{}
  for x:=0; x<DISTRIBUTION_MAX; x++ {
    rand.Seed(time.Now().UTC().UnixNano())
    me.Distribution[x] = RandKey()
    me.Network.Set(me.Distribution[x], &me.ServerNetworkNode)
  }
}

func RandKey() bt.ByteSliceKey {
  rand.Seed(time.Now().UTC().UnixNano())
  b := [16]byte{}
  for i:=0; i<16; i++ {
    b[i] = byte(rand.Intn(256))
  }
  x := bt.ByteSliceKey(b)
  return x
}

func (me *ServerNode) AddToNetwork(server *ServerNetworkNode) error {
  if server.ID == me.ID {
    return errors.New("Cannot register a node with itself")
  }
  if _, found := me.NetworkNodes[server.ID]; found {
    return errors.New("Node is already registered")
  }
  me.NetworkNodes[server.ID] = server
  for _, x := range server.Distribution { me.Network.Set(x, server);  }

  me.Network.Balance()
  return nil
}

func (me *ServerNode) RemoveFromNetwork(server *ServerNetworkNode) error {
  if server.ID == me.ID {
    return errors.New("Cannot deregister a node with itself")
  }
  if _, found := me.NetworkNodes[server.ID]; !found {
    return errors.New("Node is not registered")
  }
  me.NetworkNodes[server.ID] = server
  for _, x := range server.Distribution { me.Network.Clear(x);  }

  me.Network.Balance()
  return nil
}

func (me *ServerNode) GetNodeFor(key bt.ByteSliceKey) *ServerNetworkNode {
  found, node := me.Network.Next(key)
  if !found { _, node = me.Network.First() }
  return node.(*ServerNetworkNode)
}