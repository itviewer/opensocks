package proxy

import (
    "github.com/itviewer/opensocks/base"
    "github.com/itviewer/opensocks/common/enum"
    "net"
    "time"

    "github.com/xtaci/kcp-go/v5"
)

type KCPProxy struct {
    config   base.Config
    listener *kcp.Listener
}

func (p *KCPProxy) StartProxyServer() {
    // key := pbkdf2.Key([]byte(p.config.Key), []byte("opensocks@2022"), 4096, 32, sha1.New)
    // block, _ := kcp.NewSalsa20BlockCrypt(key)
    var err error
    p.listener, err = kcp.ListenWithOptions(p.config.ServerAddr, nil, 10, 0)

    if err == nil {
        base.Info("opensocks kcp server started on", p.config.ServerAddr)
        for {
            conn, err := p.listener.AcceptKCP()
            if err != nil {
                break
            }

            conn.SetStreamMode(false)

            conn.SetWindowSize(enum.SndWnd, enum.RcvWnd)
            conn.SetNoDelay(1, 10, 2, 1)
            conn.SetMtu(1400)
            conn.SetACKNoDelay(false)

            conn.SetReadBuffer(enum.SockBuf)
            conn.SetWriteBuffer(enum.SockBuf)
            conn.SetDSCP(46)
            conn.SetReadDeadline(time.Now().Add(time.Minute))
            conn.SetWriteDeadline(time.Now().Add(time.Minute))

            go p.Handler(conn)
        }
    }
}

func (p *KCPProxy) StopProxyServer() {
    if err := p.listener.Close(); err != nil {
        base.Error("failed to shutdown kcp server:", err)
    }
}

func (p *KCPProxy) Handler(conn net.Conn) {
    MuxHandler(conn)
}
