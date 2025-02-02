package proxy

import (
    "context"
    "fmt"
    "github.com/gorilla/websocket"
    "github.com/itviewer/opensocks/base"
    "github.com/itviewer/opensocks/common/enum"
    "github.com/itviewer/opensocks/counter"
    "io"
    "log"
    "net"
    "net/http"
    "strconv"
    "strings"
)

var _defaultHomePage = []byte(`
<!DOCTYPE html>
<html>
<head>
<title>Welcome to opensocks!</title>
</head>
<body>
<p>Welcome to opensocks!</p>
</body>
</html>`)
var _wsUpgrader = websocket.Upgrader{ReadBufferSize: enum.BufferSize, WriteBufferSize: enum.BufferSize}

type WSProxy struct {
    config base.Config
    server http.Server
}

func (p *WSProxy) StartProxyServer() {
    http.HandleFunc(enum.WSPath, func(w http.ResponseWriter, r *http.Request) {
        conn, err := _wsUpgrader.Upgrade(w, r, nil)
        if err != nil {
            base.Error("failed to upgrade WebSocket", err)
            return
        }
        p.Handler(conn.NetConn())
    })

    http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
        w.WriteHeader(200)
        w.Header().Set("Content-Type", "text/html")
        w.Header().Set("Content-Length", strconv.Itoa(len(_defaultHomePage)))
        w.Header().Set("Connection", " keep-alive")
        w.Header().Set("Accept-Ranges", "bytes")
        w.Write(_defaultHomePage)
    })

    http.HandleFunc("/ip", func(w http.ResponseWriter, req *http.Request) {
        ip := req.Header.Get("X-Forwarded-For")
        if ip == "" {
            ip = strings.Split(req.RemoteAddr, ":")[0]
        }
        resp := fmt.Sprintf("%v", ip)
        io.WriteString(w, resp)
    })

    http.HandleFunc("/stats", func(w http.ResponseWriter, req *http.Request) {
        io.WriteString(w, counter.PrintServerBytes())
    })

    base.Info("opensocks ws server started on", p.config.ServerAddr)
    p.server = http.Server{
        Addr: p.config.ServerAddr,
    }
    p.server.ListenAndServe()
}

func (p *WSProxy) StopProxyServer() {
    if err := p.server.Shutdown(context.Background()); err != nil {
        log.Printf("failed to shutdown ws server: %v", err)
    }
}

func (p *WSProxy) Handler(conn net.Conn) {
    MuxHandler(conn)
}
