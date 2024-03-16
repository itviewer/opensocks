package proxy

import (
    "context"
    "fmt"
    "github.com/gorilla/websocket"
    "github.com/itviewer/opensocks/common/enum"
    "github.com/itviewer/opensocks/config"
    utls "github.com/refraction-networking/utls"
    "github.com/xtaci/kcp-go/v5"
    "log"
    "net"
    "strings"
    "time"
)

func SetupTunnel(config config.Config) net.Conn {
    protocol := strings.ToLower(config.Protocol)
    switch protocol {
    case "ws":
        return connectWSServer(config)
    case "kcp":
        return connectKCPServer(config)
    default:
        return connectTCPServer(config)
    }
}

func connectTCPServer(config config.Config) net.Conn {
    conn, err := net.DialTimeout("tcp", config.ServerAddr, time.Duration(enum.Timeout)*time.Second)
    if err != nil {
        log.Printf("[client] failed to dial tcp server %s %v", config.ServerAddr, err)
        return nil
    }
    log.Printf("[client] tcp tunnel connected %s", config.ServerAddr)
    return conn
}

func connectKCPServer(config config.Config) net.Conn {
    // key := pbkdf2.Key([]byte(config.Key), []byte("opensocks@2022"), 4096, 32, sha1.New)
    // block, _ := kcp.NewSalsa20BlockCrypt(key)
    conn, err := kcp.DialWithOptions(config.ServerAddr, nil, 10, 0)
    if err != nil {
        log.Printf("[client] failed to dial kcp server %s %v", config.ServerAddr, err)
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

    log.Printf("[client] kcp tunnel connected %s", config.ServerAddr)

    return conn
}

func connectWSServer(config config.Config) net.Conn {
    url := fmt.Sprintf("%s://%s%s", config.Protocol, config.ServerAddr, enum.WSPath)
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
        log.Printf("[client] failed to dial websocket %s %v", url, err)
        return nil
    }
    log.Printf("[client] ws tunnel connected %s", url)
    return wsconn.NetConn()
}
