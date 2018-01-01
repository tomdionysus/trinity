package network

import (
	"bufio"
	"fmt"
	"github.com/tomdionysus/trinity/util"
	"net"
	"strconv"
	"strings"
	"time"
)

type MemcacheServer struct {
	Logger   *util.Logger
	Port     int
	Server   *TLSServer
	Listener net.Listener

	Connections map[string]net.Conn
}

func NewMemcacheServer(logger *util.Logger, port int, server *TLSServer) *MemcacheServer {
	inst := &MemcacheServer{
		Logger:      logger,
		Port:        port,
		Server:      server,
		Connections: map[string]net.Conn{},
	}
	return inst
}

func (mcs *MemcacheServer) Init() error {
	mcs.Logger.Debug("Memcache", "Init")
	return nil
}

func (mcs *MemcacheServer) Start() error {
	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", mcs.Port))
	if err != nil {
		mcs.Logger.Error("Memcache", "Cannot bind to port [%d], shutting down", mcs.Port)
		return err
	}
	mcs.Listener = listener
	mcs.Logger.Info("Memcache", "Listening on port [%d]", mcs.Port)
	go func() {
		for {
			// Wait for a connection.
			conn, err := listener.Accept()
			if err != nil {
				mcs.Logger.Info("Memcache", "Closed Listener")
				break
			} else {
				addr := conn.RemoteAddr().String()
				mcs.Logger.Info("Memcache", "Incoming Connection from [%s]", addr)
				go mcs.handleConnection(addr, conn)
			}
		}
	}()
	return nil
}

func (mcs *MemcacheServer) Stop() {
	// Listener
	if mcs.Listener != nil {
		mcs.Listener.Close()
		mcs.Listener = nil
	}
	// Close all connections
	for addr, conn := range mcs.Connections {
		mcs.Logger.Debug("Memcache", "Force closing Connection [%s]", addr)
		conn.Close()
	}
}

// Private

func (mcs *MemcacheServer) handleConnection(addr string, conn net.Conn) {
	mcs.Connections[addr] = conn

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	mcs.Logger.Debug("Memcache", "[%s] -> Connected", addr)
	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			if strings.HasSuffix(err.Error(), "use of closed network connection") {
				mcs.Logger.Debug("Memcache", "[%s] -> Disconnected", addr)
				break
			}
			mcs.Logger.Error("Memcache", "[%s] -> Error: %s", addr, err.Error())
		}
		cmds := strings.Split(strings.Trim(input, " \n\r"), " ")
		if mcs.handleCommand(addr, reader, writer, cmds) {
			break
		}
	}

	conn.Close()
	delete(mcs.Connections, addr)
}

func (mcs *MemcacheServer) handleCommand(addr string, reader *bufio.Reader, writer *bufio.Writer, args []string) bool {

	if len(args) == 0 {
		writer.WriteString("ERROR\r\n")
		writer.Flush()
		return false
	}

	switch args[0] {
	case "exit":
		writer.WriteString("END\r\n")
		writer.Flush()
		return true
	case "set":
		mcs.handleSet(addr, reader, writer, args)
		return false
	case "get":
		mcs.handleGet(addr, reader, writer, args)
		return false
	case "delete":
		mcs.handleDelete(addr, reader, writer, args)
		return false
	default:
		writer.WriteString("ERROR\r\n")
		writer.Flush()
	}

	return false
}

func (mcs *MemcacheServer) handleSet(addr string, reader *bufio.Reader, writer *bufio.Writer, args []string) {
	if len(args) > 6 || len(args) < 5 {
		writer.WriteString("ERROR\r\n")
		writer.Flush()
		return
	}
	// args[1] key
	// args[2] flags
	// args[3] exptime
	// args[4] bytes
	// args[5] noreply

	mcs.Logger.Debug("Memcache", "[%s] -> Set %s", addr, args)

	expirytime, err := strconv.Atoi(args[3])
	if err != nil {
		writer.WriteString("SERVER_ERROR\r\n")
		writer.Flush()
		return
	}

	flags, err := strconv.Atoi(args[2])
	flags = flags & 0xFFFF
	if err != nil {
		writer.WriteString("SERVER_ERROR\r\n")
		writer.Flush()
		return
	}

	bytes, err := strconv.Atoi(args[4])
	if err != nil {
		writer.WriteString("SERVER_ERROR\r\n")
		writer.Flush()
		return
	}

	var buf []byte = make([]byte, bytes, bytes)
	n, err := reader.Read(buf)
	if err != nil || n != len(buf) {
		writer.WriteString("SERVER_ERROR\r\n")
		writer.Flush()
		return
	}

	_, err = reader.ReadString('\n')
	if err != nil {
		writer.WriteString("SERVER_ERROR\r\n")
		writer.Flush()
		return
	}

	var expparam *time.Time = nil
	if expirytime != 0 {
		expiry := time.Now().UTC().Add(time.Duration(expirytime) * time.Second)
		expparam = &expiry
	}
	mcs.Server.SetKey(args[1], buf[:], int16(flags), expparam)
	writer.WriteString("STORED\r\n")
	writer.Flush()
}

func (mcs *MemcacheServer) handleGet(addr string, reader *bufio.Reader, writer *bufio.Writer, args []string) {
	if len(args) > 2 {
		writer.WriteString("ERROR\r\n")
		writer.Flush()
		return
	}
	mcs.Logger.Debug("Memcache", "[%s] -> Get Key %s", addr, args[1])
	value, flags, found := mcs.Server.GetKey(args[1])
	if found {
		mcs.Logger.Debug("Memcache", "[%s] -> Found", addr)
		writer.WriteString(fmt.Sprintf("VALUE %s %d %d\r\n", args[1], flags, len(value)))
		writer.Write(value)
		writer.Write([]byte{13, 10})
	}
	writer.WriteString("END\r\n")
	writer.Flush()
}

func (mcs *MemcacheServer) handleDelete(addr string, reader *bufio.Reader, writer *bufio.Writer, args []string) {
	if len(args) > 3 {
		writer.WriteString("ERROR\r\n")
		writer.Flush()
		return
	}
	mcs.Logger.Debug("Memcache", "[%s] -> Delete Key %s", addr, args[1])
	found := mcs.Server.DeleteKey(args[1])
	if found {
		mcs.Logger.Debug("Memcache", "[%s] -> Found", addr)
		writer.WriteString("DELETED\r\n")
		writer.Flush()
	} else {
		mcs.Logger.Debug("Memcache", "[%s] -> Not Found", addr)
		writer.WriteString("NOT_FOUND\r\n")
		writer.Flush()
	}
}
