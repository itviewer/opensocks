package proxy

import (
    "bufio"
    "github.com/itviewer/opensocks/base"
    "github.com/itviewer/opensocks/common/handshake"
    "github.com/itviewer/opensocks/common/pool"
    "net"
    "time"

    "github.com/itviewer/opensocks/codec"
    "github.com/itviewer/opensocks/common/enum"
    "github.com/itviewer/opensocks/counter"
    "github.com/xtaci/smux"
)

func MuxHandler(w net.Conn) {
    defer w.Close()
    smuxConfig := smux.DefaultConfig()
    smuxConfig.Version = enum.SmuxVer
    smuxConfig.MaxReceiveBuffer = enum.SmuxBuf
    smuxConfig.MaxStreamBuffer = enum.StreamBuf
    session, err := smux.Server(w, smuxConfig)
    if err != nil {
        base.Error("failed to initialise smux session:", err)
        return
    }
    defer session.Close()
    for {
        stream, err := session.AcceptStream()
        if err != nil {
            base.Error("failed to accept steam", err)
            break
        }
        go func() {
            defer stream.Close()
            reader := bufio.NewReader(stream)
            // handshake
            ok, req := handshake.ReadHelloRequest(reader)
            if !ok {
                return
            }
            base.Debug("dial to server", req.Network, req.Host, req.Port)
            conn, err := net.DialTimeout(req.Network, net.JoinHostPort(req.Host, req.Port), time.Duration(enum.Timeout)*time.Second)
            if err != nil {
                base.Debug("failed to dial server", err)
                return
            }
            // forwarding data
            go toServer(reader, conn)
            toClient(stream, conn)
        }()
    }
}

func toClient(stream net.Conn, conn net.Conn) {
    defer conn.Close()
    buffer := pool.Get()
    defer pool.Put(buffer)
    for {
        n, err := conn.Read(buffer)
        if err != nil {
            break
        }
        b := buffer[:n]
        b = codec.EncodeData(b)
        _, err = stream.Write(b)
        if err != nil {
            break
        }
        counter.IncrWrittenBytes(n)
    }
}

func toServer(streamReader *bufio.Reader, conn net.Conn) {
    defer conn.Close()
    buffer := pool.Get()
    defer pool.Put(buffer)
    for {
        n, err := streamReader.Read(buffer)
        if err != nil {
            break
        }
        b := buffer[:n]
        b, err = codec.DecodeData(b)
        if err != nil {
            base.Debug("failed to decode:", err)
            break
        }
        _, err = conn.Write(b)
        if err != nil {
            break
        }
        counter.IncrReadBytes(int(n))
    }
}
