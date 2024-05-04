package socks5

import (
    "github.com/itviewer/opensocks/base"
    "github.com/itviewer/opensocks/client/proxy"
    "github.com/itviewer/opensocks/codec"
    "github.com/itviewer/opensocks/common/enum"
    "github.com/itviewer/opensocks/common/handshake"
    "github.com/itviewer/opensocks/common/pool"
    "github.com/itviewer/opensocks/counter"
    "github.com/xtaci/smux"
    "io"
    "net"
    "sync"
)

// TCPProxy The tcp proxy struct
type TCPProxy struct {
    Session *smux.Session
    Lock    sync.Mutex
}

// Proxy is a function to proxy data
func (t *TCPProxy) Proxy(tcpConn net.Conn, req []byte) {
    host, port := parseAddr(req)
    if host == "" || port == "" {
        return
    }
    // bypass private ip
    if base.Cfg.Bypass {
        ip := net.ParseIP(host)
        if ip != nil && ip.IsPrivate() {
            directProxy(tcpConn, host, port)
            return
        }
    }
    t.Lock.Lock()
    if t.Session == nil {
        var err error
        xconn := proxy.SetupTunnel()
        if xconn == nil {
            t.Lock.Unlock()
            resp(tcpConn, enum.ConnectionRefused)
            return
        }
        t.Session, err = newMuxSession(xconn)
        if err != nil || t.Session == nil {
            t.Lock.Unlock()
            base.Error("failed to initialize a new smux connection:", err)
            resp(tcpConn, enum.ConnectionRefused)
            return
        }
    }
    t.Lock.Unlock()
    // create a new stream
    stream, err := t.Session.Open()
    if err != nil {
        t.Session = nil
        base.Debug("failed to open smux session:", err)
        resp(tcpConn, enum.ConnectionRefused)
        return
    }
    ok := handshake.HelloToTarget(stream, "tcp", host, port, base.Cfg.Key, base.Cfg.Obfs)
    if !ok {
        t.Session = nil
        base.Error("failed to handshake to", host)
        resp(tcpConn, enum.ConnectionRefused)
        return
    }
    resp(tcpConn, enum.SuccessReply)
    go t.toServer(stream, tcpConn)
    t.toClient(stream, tcpConn)
}

// toServer is a goroutine to copy data from client to server
func (t *TCPProxy) toServer(stream io.ReadWriteCloser, tcpconn net.Conn) {
    defer stream.Close()
    defer tcpconn.Close()
    buffer := pool.Get()
    defer pool.Put(buffer)
    for {
        n, err := tcpconn.Read(buffer)
        if err != nil {
            break
        }
        b := buffer[:n]
        b = codec.EncodeData(b)
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
    buffer := pool.Get()
    defer pool.Put(buffer)
    for {
        n, err := stream.Read(buffer)
        if err != nil {
            break
        }
        b := buffer[:n]
        b, err = codec.DecodeData(b)
        if err != nil {
            base.Debug("failed to decode:", err)
            break
        }
        _, err = tcpconn.Write(b)
        if err != nil {
            break
        }
        counter.IncrReadBytes(n)
    }
}
