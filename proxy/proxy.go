package proxy

import (
	"net"
	"strings"

	"github.com/net-byte/opensocks/config"
)

type Proxy interface {
	ConnectServer() net.Conn
	StartServer()
	StopServer()
	Handler(conn net.Conn)
}

func NewProxy(config config.Config) Proxy {
	protocol := strings.ToLower(config.Protocol)
	switch protocol {
	case "tcp":
		return &TCPProxy{config: config}
	case "kcp":
		return &KCPProxy{config: config}
	default:
		return &WSProxy{config: config}
	}
}
