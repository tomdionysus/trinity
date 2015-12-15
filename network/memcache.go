package network

import (
  "net"
  "github.com/tomdionysus/trinity/util"
  "github.com/tomdionysus/trinity/kvstore"
  "fmt"
  "bufio"
  "strings"
  "strconv"
  "time"
)

type MemcacheServer struct {
  Logger *util.Logger
  Port int
  KVStore *kvstore.KVStore
  Listener net.Listener

  Connections map[string]net.Conn
}

func NewMemcacheServer(logger *util.Logger, port int, kv *kvstore.KVStore) *MemcacheServer {
  inst := &MemcacheServer{
    Logger: logger,
    Port: port,
    KVStore: kv,
    Connections: map[string]net.Conn{},
  }
  return inst
}

func (me *MemcacheServer) Init() error {
  me.Logger.Debug("Memcache","Init")
  return nil
}


func (me *MemcacheServer) Start() error {
  listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d",me.Port))
  if err != nil { 
    me.Logger.Error("Memcache","Cannot bind to port [%d], shutting down", me.Port)
    return err
  }
  me.Listener = listener
  me.Logger.Debug("Memcache","Listening on port [%d]", me.Port)
  go func() {
    for {
      // Wait for a connection.
      conn, err := listener.Accept()
      if err != nil {
        me.Logger.Info("Memcache","Incoming Connection Failed: %s",err.Error())
      } else {
        addr := conn.RemoteAddr().String()
        me.Logger.Info("Memcache","Incoming Connection from [%s]",addr)
        go me.handleConnection(addr, conn)
      }
    }
  }()
  return nil
}

func (me *MemcacheServer) Stop() {
  // Listener
  if me.Listener!=nil {
    me.Listener.Close()
    me.Listener = nil
  }
  // Close all connections
  for addr, conn := range me.Connections {
    me.Logger.Debug("Memcache", "Force closing Connection [%s]", addr)
    conn.Close()
  }
}

// Private

func (me *MemcacheServer) handleConnection(addr string, conn net.Conn) {
  me.Connections[addr] = conn

  reader := bufio.NewReader(conn)
  writer := bufio.NewWriter(conn)

  me.Logger.Debug("Memcache", "[%s] -> Connected",addr)
  for {
    input, err := reader.ReadString('\n')
    if err != nil {
      me.Logger.Error("Memcache", "[%s] -> Error: %s", addr, err.Error())
    }
    cmds := strings.Split(strings.Trim(input," \n\r")," ")
    if me.handleCommand(addr, reader, writer, cmds) { break }
  }

  conn.Close()
  delete(me.Connections, addr)
}

func (me *MemcacheServer) handleCommand(addr string, reader *bufio.Reader, writer *bufio.Writer, args []string) bool {
  
  if len(args) == 0 {
    writer.WriteString("ERROR\r\n"); writer.Flush()
    return false
  }

  switch args[0] {
    case "exit":
      writer.WriteString("END\r\n"); writer.Flush()
      return true
    case "set":
      me.handleSet(addr, reader, writer, args)
      return false
    case "get":
      me.handleGet(addr, reader, writer, args)
      return false
    case "delete":
      me.handleDelete(addr, reader, writer, args)
      return false
    default:
      writer.WriteString("ERROR\r\n"); writer.Flush()
      return false
  }

  return false
}

func (me *MemcacheServer) handleSet(addr string, reader *bufio.Reader, writer *bufio.Writer, args []string) {
  if len(args)>6 || len(args)<5 {
    writer.WriteString("ERROR\r\n"); writer.Flush()
    return
  }
  // args[1] key
  // args[2] flags
  // args[3] exptime
  // args[4] bytes
  // args[5] noreply

  me.Logger.Debug("Memcache", "[%s] -> Set %s", addr, args)

  bytes, err := strconv.Atoi(args[4])
  if err!=nil { writer.WriteString("SERVER_ERROR\r\n"); writer.Flush(); return }

  expirytime, err := strconv.Atoi(args[3])
  if err!=nil { writer.WriteString("SERVER_ERROR\r\n"); writer.Flush(); return }

  var buf []byte = make([]byte,bytes,bytes)
  n, err := reader.Read(buf)
  if err!=nil || n!=len(buf) {
    writer.WriteString("SERVER_ERROR\r\n"); writer.Flush()
    return
  }

  _, err = reader.ReadString('\n')
  if err!=nil {
    writer.WriteString("SERVER_ERROR\r\n"); writer.Flush()
    return
  }

  var expparam *time.Time = nil
  if expirytime!=0 {
    expiry := time.Now().UTC().Add(time.Duration(expirytime)*time.Second)
    expparam = &expiry 
  }
  me.KVStore.Set(args[1], buf[:], expparam)
  writer.WriteString("STORED\r\n") 
  writer.Flush()
}

func (me *MemcacheServer) handleGet(addr string, reader *bufio.Reader, writer *bufio.Writer, args []string) {
  if len(args)>2 {
    writer.WriteString("ERROR\r\n"); writer.Flush()
    return
  }
  me.Logger.Debug("Memcache", "[%s] -> Get Key %s", addr, args[1])
  value, found := me.KVStore.Get(args[1])
  if found {
    me.Logger.Debug("Memcache", "[%s] -> Found", addr)
    writer.WriteString(fmt.Sprintf("VALUE %s %s %d\r\n", args[1], "0", len(value)))
    writer.Write(value)
    writer.Write([]byte{ 13, 10 })
  }
  writer.WriteString("END\r\n"); writer.Flush()
}

func (me *MemcacheServer) handleDelete(addr string, reader *bufio.Reader, writer *bufio.Writer, args []string) {
  if len(args)>3 {
    writer.WriteString("ERROR\r\n"); writer.Flush()
    return
  }
  me.Logger.Debug("Memcache", "[%s] -> Delete Key %s", addr, args[1])
  found := me.KVStore.Delete(args[1])
  if found {
    me.Logger.Debug("Memcache", "[%s] -> Found", addr)
    writer.WriteString("DELETED\r\n"); writer.Flush()
  } else {
    me.Logger.Debug("Memcache", "[%s] -> Not Found", addr)
    writer.WriteString("NOT_FOUND\r\n"); writer.Flush()
  }
}
