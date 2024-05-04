package proxy

import (
    "github.com/itviewer/opensocks/base"
    "net"
    "strings"
)

type Proxy interface {
    StartProxyServer()
    StopProxyServer()
    Handler(conn net.Conn)
}

func NewProxy() Proxy {
    protocol := strings.ToLower(base.Cfg.Protocol)
    switch protocol {
    case "tcp":
        return &TCPProxy{}
    case "kcp":
        return &KCPProxy{}
    case "ws":
        return &WSProxy{}
    default:
        return nil // 拼写错误
    }
}
