package proxy

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/net-byte/opensocks/common/enum"
	"github.com/net-byte/opensocks/config"
	"github.com/net-byte/opensocks/counter"
	utls "github.com/refraction-networking/utls"
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
	config config.Config
	server http.Server
}

func (p *WSProxy) ConnectServer() net.Conn {
	url := fmt.Sprintf("%s://%s%s", p.config.Protocol, p.config.ServerAddr, enum.WSPath)
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
	log.Printf("[client] ws server connected %s", url)
	return wsconn.UnderlyingConn()
}

func (p *WSProxy) StartServer() {
	http.HandleFunc(enum.WSPath, func(w http.ResponseWriter, r *http.Request) {
		conn, err := _wsUpgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("[server] failed to upgrade http %v", err)
			return
		}
		p.Handler(conn.UnderlyingConn())
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

	log.Printf("opensocks ws server started on %s", p.config.ServerAddr)
	p.server = http.Server{
		Addr: p.config.ServerAddr,
	}
	p.server.ListenAndServe()
}

func (p *WSProxy) StopServer() {
	if err := p.server.Shutdown(context.Background()); err != nil {
		log.Printf("failed to shutdown ws server: %v", err)
	}
}

func (p *WSProxy) Handler(conn net.Conn) {
	MuxHandler(conn, p.config)
}
