package proxy

import (
	"log"
	"net"
	"time"

	"github.com/net-byte/opensocks/common/enum"
	"github.com/net-byte/opensocks/config"
)

type TCPProxy struct {
	config   config.Config
	listener net.Listener
}

func (p *TCPProxy) ConnectServer() net.Conn {
	c, err := net.DialTimeout("tcp", p.config.ServerAddr, time.Duration(enum.Timeout)*time.Second)
	if err != nil {
		log.Printf("[client] failed to dial tcp server %s %v", p.config.ServerAddr, err)
		return nil
	}
	log.Printf("[client] tcp server connected %s", p.config.ServerAddr)
	return c
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
