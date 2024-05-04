package proxy

import (
    "github.com/itviewer/opensocks/base"
    "net"
)

type TCPProxy struct {
    listener net.Listener
}

func (p *TCPProxy) StartProxyServer() {
    var err error
    if p.listener, err = net.Listen("tcp", base.Cfg.ServerAddr); err == nil {
        base.Info("opensocks tcp server started on", base.Cfg.ServerAddr)
        for {
            conn, err := p.listener.Accept()
            if err != nil {
                base.Error(err)
                break
            }
            go p.Handler(conn)
        }
    }
}

func (p *TCPProxy) StopProxyServer() {
    if err := p.listener.Close(); err != nil {
        base.Error("failed to shutdown tcp server:", err)
    }
}

func (p *TCPProxy) Handler(conn net.Conn) {
    MuxHandler(conn)
}
