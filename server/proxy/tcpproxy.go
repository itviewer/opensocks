package proxy

import (
    "github.com/itviewer/opensocks/config"
    "log"
    "net"
)

type TCPProxy struct {
    config   config.Config
    listener net.Listener
}

func (p *TCPProxy) StartServer() {
    var err error
    if p.listener, err = net.Listen("tcp", p.config.ServerAddr); err == nil {
        log.Printf("opensocks tcp server started on %s", p.config.ServerAddr)
        for {
            conn, err := p.listener.Accept()
            if err != nil {
                break
            }
            go p.Handler(conn)
        }
    }
}

func (p *TCPProxy) StopServer() {
    if err := p.listener.Close(); err != nil {
        log.Printf("failed to shutdown tcp server: %v", err)
    }
}

func (p *TCPProxy) Handler(conn net.Conn) {
    MuxHandler(conn, p.config)
}
