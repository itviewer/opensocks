package socks5

import (
    "bytes"
    "github.com/itviewer/opensocks/base"
    "github.com/itviewer/opensocks/client/proxy"
    "github.com/itviewer/opensocks/common/handshake"
    "github.com/itviewer/opensocks/common/pool"
    "io"
    "net"
    "strconv"
    "sync"

    "github.com/itviewer/opensocks/codec"
    "github.com/itviewer/opensocks/counter"
    "github.com/xtaci/smux"
)

type UDPServer struct {
    TCPProxy  *TCPProxy
    UDPConn   *net.UDPConn
    headerMap sync.Map
    streamMap sync.Map
    Session   *smux.Session
    Lock      sync.Mutex
}

// Start the UDP server
func (u *UDPServer) Start() *net.UDPConn {
    udpAddr, _ := net.ResolveUDPAddr("udp", base.Cfg.LocalAddr)
    var err error
    u.UDPConn, err = net.ListenUDP("udp", udpAddr)
    if err != nil {
        base.Error("failed to listen udp", err)
        return nil
    }
    base.Info("udp server started on", base.Cfg.LocalAddr)
    go u.toServer()
    return u.UDPConn
}

// toServer handle the udp packet from client
func (u *UDPServer) toServer() {
    defer u.UDPConn.Close()
    buf := pool.Get()
    defer pool.Put(buf)
    for {
        n, cliAddr, err := u.UDPConn.ReadFromUDP(buf)
        if err != nil {
            break
        }
        b := buf[:n]
        dstAddr, header, data := parseUDPData(b)
        if dstAddr == nil || header == nil || data == nil {
            continue
        }
        key := cliAddr.String()
        var stream io.ReadWriteCloser
        if value, ok := u.streamMap.Load(key); !ok {
            u.Lock.Lock()
            if u.Session == nil {
                var err error
                xconn := proxy.SetupTunnel()
                if xconn == nil {
                    u.Lock.Unlock()
                    continue
                }
                u.Session, err = newMuxSession(xconn)
                if err != nil || u.Session == nil {
                    base.Error(err)
                    u.Lock.Unlock()
                    continue
                }
            }
            u.Lock.Unlock()
            stream, err = u.Session.Open()
            if err != nil {
                u.Session = nil
                base.Error("failed to open smux session:", err)
                continue
            }
            ok := handshake.HelloToTarget(stream, "udp", dstAddr.IP.String(), strconv.Itoa(dstAddr.Port), base.Cfg.Key, base.Cfg.Obfs)
            if !ok {
                u.Session = nil
                base.Error("failed to handshake to", dstAddr.IP.String())
                continue
            }
            u.streamMap.Store(key, stream)
            u.headerMap.Store(key, header)
            go u.toClient(stream, cliAddr)
        } else {
            stream = value.(io.ReadWriteCloser)
        }
        data = codec.EncodeData(data)
        stream.Write(data)
        counter.IncrWrittenBytes(n)
    }
}

// toClient handle the udp packet from server
func (u *UDPServer) toClient(stream io.ReadWriteCloser, cliAddr *net.UDPAddr) {
    key := cliAddr.String()
    buffer := pool.Get()
    defer pool.Put(buffer)
    defer stream.Close()
    for {
        n, err := stream.Read(buffer)
        if err != nil {
            break
        }
        if header, ok := u.headerMap.Load(key); ok {
            b := buffer[:n]
            b, err = codec.DecodeData(b)
            if err != nil {
                base.Error("failed to decode:", err)
                break
            }
            var data bytes.Buffer
            data.Write(header.([]byte))
            data.Write(b)
            _, err = u.UDPConn.WriteToUDP(data.Bytes(), cliAddr)
            if err != nil {
                break
            }
            counter.IncrReadBytes(n)
        }
    }
    u.headerMap.Delete(key)
    u.streamMap.Delete(key)
}
