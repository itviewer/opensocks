package proxy

import (
    "bufio"
    "github.com/itviewer/opensocks/common/handshake"
    "github.com/itviewer/opensocks/common/pool"
    "log"
    "net"
    "time"

    "github.com/itviewer/opensocks/codec"
    "github.com/itviewer/opensocks/common/enum"
    "github.com/itviewer/opensocks/common/util"
    "github.com/itviewer/opensocks/config"
    "github.com/itviewer/opensocks/counter"
    "github.com/xtaci/smux"
)

func MuxHandler(w net.Conn, config config.Config) {
    defer w.Close()
    smuxConfig := smux.DefaultConfig()
    smuxConfig.Version = enum.SmuxVer
    smuxConfig.MaxReceiveBuffer = enum.SmuxBuf
    smuxConfig.MaxStreamBuffer = enum.StreamBuf
    session, err := smux.Server(w, smuxConfig)
    if err != nil {
        log.Printf("[server] failed to initialise smux session: %s", err)
        return
    }
    defer session.Close()
    for {
        stream, err := session.AcceptStream()
        if err != nil {
            util.PrintLog(config.Verbose, "[server] failed to accept steam %v", err)
            break
        }
        go func() {
            defer stream.Close()
            reader := bufio.NewReader(stream)
            // handshake
            ok, req := handshake.ReadAddressingRequest(config, reader)
            if !ok {
                return
            }
            util.PrintLog(config.Verbose, "[server] dial to server %v", req.Network, req.Host, req.Port)
            conn, err := net.DialTimeout(req.Network, net.JoinHostPort(req.Host, req.Port), time.Duration(enum.Timeout)*time.Second)
            if err != nil {
                util.PrintLog(config.Verbose, "[server] failed to dial server %v", err)
                return
            }
            // forwarding data
            go toServer(config, reader, conn)
            toClient(config, stream, conn)
        }()
    }
}

func toClient(config config.Config, stream net.Conn, conn net.Conn) {
    defer conn.Close()
    buffer := pool.Get()
    defer pool.Put(buffer)
    for {
        n, err := conn.Read(buffer)
        if err != nil {
            break
        }
        b := buffer[:n]
        b = codec.EncodeData(b, config)
        _, err = stream.Write(b)
        if err != nil {
            break
        }
        counter.IncrWrittenBytes(n)
    }
}

func toServer(config config.Config, streamReader *bufio.Reader, conn net.Conn) {
    defer conn.Close()
    buffer := pool.Get()
    defer pool.Put(buffer)
    for {
        n, err := streamReader.Read(buffer)
        if err != nil {
            break
        }
        b := buffer[:n]
        b, err = codec.DecodeData(b, config)
        if err != nil {
            util.PrintLog(config.Verbose, "failed to decode:%v", err)
            break
        }
        _, err = conn.Write(b)
        if err != nil {
            break
        }
        counter.IncrReadBytes(int(n))
    }
}
