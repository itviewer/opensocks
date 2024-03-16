package proxy

import (
    "github.com/itviewer/opensocks/common/enum"
    "log"
    "net"
    "time"

    "github.com/itviewer/opensocks/config"
    "github.com/xtaci/kcp-go/v5"
)

type KCPProxy struct {
    config   config.Config
    listener *kcp.Listener
}

func (p *KCPProxy) StartServer() {
    // key := pbkdf2.Key([]byte(p.config.Key), []byte("opensocks@2022"), 4096, 32, sha1.New)
    // block, _ := kcp.NewSalsa20BlockCrypt(key)
    var err error
    p.listener, err = kcp.ListenWithOptions(p.config.ServerAddr, nil, 10, 0)

    if err == nil {
        log.Printf("opensocks kcp server started on %s", p.config.ServerAddr)
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

func (p *KCPProxy) StopServer() {
    if err := p.listener.Close(); err != nil {
        log.Printf("failed to shutdown kcp server: %v", err)
    }
}

func (p *KCPProxy) Handler(conn net.Conn) {
    MuxHandler(conn, p.config)
}
