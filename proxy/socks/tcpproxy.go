package socks

import (
	"io"
	"log"
	"net"
	"sync"

	"github.com/net-byte/opensocks/codec"
	"github.com/net-byte/opensocks/common/enum"
	"github.com/net-byte/opensocks/common/pool"
	"github.com/net-byte/opensocks/common/util"
	"github.com/net-byte/opensocks/config"
	"github.com/net-byte/opensocks/counter"
	"github.com/net-byte/opensocks/proxy"
	"github.com/xtaci/smux"
)

// The tcp proxy struct
type TCPProxy struct {
	Config  config.Config
	Session *smux.Session
	Lock    sync.Mutex
}

// Proxy is a function to proxy data
func (t *TCPProxy) Proxy(conn net.Conn, data []byte) {
	host, port := parseAddr(data)
	if host == "" || port == "" {
		return
	}
	// bypass private ip
	if t.Config.Bypass && net.ParseIP(host) != nil && net.ParseIP(host).IsPrivate() {
		directProxy(conn, host, port, t.Config)
		return
	}
	t.Lock.Lock()
	if t.Session == nil {
		var err error
		xconn := proxy.NewProxy(t.Config).ConnectServer()
		if xconn == nil {
			t.Lock.Unlock()
			resp(conn, enum.ConnectionRefused)
			return
		}
		t.Session, err = newMuxSession(xconn)
		if err != nil || t.Session == nil {
			t.Lock.Unlock()
			util.PrintLog(t.Config.Verbose, "failed to open client:%v", err)
			resp(conn, enum.ConnectionRefused)
			return
		}
	}
	t.Lock.Unlock()
	stream, err := t.Session.Open()
	if err != nil {
		t.Session = nil
		util.PrintLog(t.Config.Verbose, "failed to open session:%v", err)
		resp(conn, enum.ConnectionRefused)
		return
	}
	ok := proxy.ClientHandshake(stream, "tcp", host, port, t.Config.Key, t.Config.Obfs)
	if !ok {
		t.Session = nil
		log.Println("[tcp] failed to handshake")
		resp(conn, enum.ConnectionRefused)
		return
	}
	resp(conn, enum.SuccessReply)
	go t.toServer(stream, conn)
	t.toClient(stream, conn)
}

// toServer is a goroutine to copy data from client to server
func (t *TCPProxy) toServer(stream io.ReadWriteCloser, tcpconn net.Conn) {
	defer stream.Close()
	defer tcpconn.Close()
	buffer := pool.BytePool.Get()
	defer pool.BytePool.Put(buffer)
	for {
		n, err := tcpconn.Read(buffer)
		if err != nil {
			break
		}
		b := buffer[:n]
		b = codec.EncodeData(b, t.Config)
		_, err = stream.Write(b)
		if err != nil {
			break
		}
		counter.IncrWrittenBytes(n)
	}
}

// toClient is a goroutine to copy data from server to client
func (t *TCPProxy) toClient(stream io.ReadWriteCloser, tcpconn net.Conn) {
	defer stream.Close()
	defer tcpconn.Close()
	buffer := pool.BytePool.Get()
	defer pool.BytePool.Put(buffer)
	for {
		n, err := stream.Read(buffer)
		if err != nil {
			break
		}
		b := buffer[:n]
		b, err = codec.DecodeData(b, t.Config)
		if err != nil {
			util.PrintLog(t.Config.Verbose, "failed to decode:%v", err)
			break
		}
		_, err = tcpconn.Write(b)
		if err != nil {
			break
		}
		counter.IncrReadBytes(n)
	}
}
