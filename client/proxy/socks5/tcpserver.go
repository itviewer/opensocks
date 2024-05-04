package socks5

import (
    "github.com/itviewer/opensocks/base"
    "net"

    "github.com/itviewer/opensocks/common/enum"
    "github.com/itviewer/opensocks/common/pool"
)

type TCPServer struct {
    TCPProxy *TCPProxy
    UDPProxy *UDPProxy
    UDPConn  *net.UDPConn
    Listener net.Listener
}

func (t *TCPServer) Start() {
    var err error
    t.Listener, err = net.Listen("tcp", base.Cfg.LocalAddr)
    if err != nil {
        base.Error("failed to listen tcp server", err)
        return
    }
    base.Info("tcp server started on", base.Cfg.LocalAddr)
    for {
        tcpConn, err := t.Listener.Accept()
        if err != nil {
            base.Error("tcp server", err)
            break
        }
        go t.handler(tcpConn)
    }
}

// handler handles the tcp connection
func (t *TCPServer) handler(tcpConn net.Conn) {
    if !t.checkVersion(tcpConn) {
        tcpConn.Close()
        return
    }
    // no auth
    respNoAuth(tcpConn)
    t.cmd(tcpConn)
}

// checkVersion checks the version
func (t *TCPServer) checkVersion(tcpConn net.Conn) bool {
    buf := pool.Get()
    defer pool.Put(buf)
    n, err := tcpConn.Read(buf[0:])
    if err != nil || n == 0 {
        return false
    }
    b := buf[0:n]
    if b[0] != enum.Socks5Version {
        resp(tcpConn, enum.ConnectionRefused)
        return false
    }
    return true
}

// cmd handles the command
func (t *TCPServer) cmd(tcpConn net.Conn) {
    buf := pool.Get()
    defer pool.Put(buf)
    n, err := tcpConn.Read(buf[0:])
    if err != nil || n == 0 {
        return
    }
    b := buf[0:n]
    switch b[1] {
    case enum.ConnectCommand:
        t.TCPProxy.Proxy(tcpConn, b)
        return
    case enum.AssociateCommand:
        t.UDPProxy.Proxy(tcpConn, t.UDPConn)
        return
    // case enum.BindCommand:
    //     resp(tcpConn, enum.CommandNotSupported)
    //     return
    default:
        resp(tcpConn, enum.CommandNotSupported)
        return
    }
}
