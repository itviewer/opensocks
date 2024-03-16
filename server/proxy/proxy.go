package proxy

import (
    "net"
    "strings"

    "github.com/itviewer/opensocks/config"
)

type Proxy interface {
    StartServer()
    StopServer()
    Handler(conn net.Conn)
}

func NewProxy(config config.Config) Proxy {
    protocol := strings.ToLower(config.Protocol)
    switch protocol {
    case "ws":
        return &WSProxy{config: config}
    case "kcp":
        return &KCPProxy{config: config}
    default:
        return &TCPProxy{config: config}
    }
}
