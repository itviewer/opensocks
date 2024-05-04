package proxy

import (
    "context"
    "fmt"
    "github.com/gorilla/websocket"
    "github.com/itviewer/opensocks/base"
    "github.com/itviewer/opensocks/common/enum"
    utls "github.com/refraction-networking/utls"
    "github.com/xtaci/kcp-go/v5"
    "net"
    "strings"
    "time"
)

func SetupTunnel() net.Conn {
    protocol := strings.ToLower(base.Cfg.Protocol)
    switch protocol {
    case "tcp":
        return connectTCPServer()
    case "kcp":
        return connectKCPServer()
    case "ws":
        return connectWSServer()
    default:
        return nil // 拼写错误
    }
}

func connectTCPServer() net.Conn {
    conn, err := net.DialTimeout("tcp", base.Cfg.ServerAddr, time.Duration(enum.Timeout)*time.Second)
    if err != nil {
        base.Error("failed to dial tcp server", base.Cfg.ServerAddr)
        return nil
    }
    base.Info("tcp tunnel connected", base.Cfg.ServerAddr)
    return conn
}

func connectKCPServer() net.Conn {
    // key := pbkdf2.Key([]byte(config.Key), []byte("opensocks@2022"), 4096, 32, sha1.New)
    // block, _ := kcp.NewSalsa20BlockCrypt(key)
    conn, err := kcp.DialWithOptions(base.Cfg.ServerAddr, nil, 10, 0)
    if err != nil {
        base.Error("failed to dial kcp server", base.Cfg.ServerAddr)
        return nil
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

    base.Info("kcp tunnel connected", base.Cfg.ServerAddr)

    return conn
}

func connectWSServer() net.Conn {
    url := fmt.Sprintf("%s://%s%s", base.Cfg.Protocol, base.Cfg.ServerAddr, enum.WSPath)
    dialer := &websocket.Dialer{
        NetDial: func(network, addr string) (net.Conn, error) {
            return net.DialTimeout(network, addr, time.Duration(enum.Timeout)*time.Second)
        },
        ReadBufferSize:   enum.BufferSize,
        WriteBufferSize:  enum.BufferSize,
        HandshakeTimeout: time.Duration(enum.Timeout) * time.Second,
    }
    dialer.NetDialTLSContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
        netDialer := &net.Dialer{}
        netConn, err := netDialer.DialContext(ctx, network, addr)
        if err != nil {
            return nil, err
        }
        // must use fingerprint with no alpn since the websocket must be handled with http/1.1
        uTLSConfig := &utls.Config{NextProtos: []string{"http/1.1"}, ServerName: strings.Split(addr, ":")[0]}
        return utls.UClient(netConn, uTLSConfig, utls.HelloRandomizedNoALPN), nil
    }
    wsconn, _, err := dialer.Dial(url, nil)
    if err != nil {
        base.Error("failed to dial websocket", url)
        return nil
    }
    base.Info("ws tunnel connected", url)
    return wsconn.NetConn()
}
