package client

import (
	"context"
	"log"
	"net/http"
	"net/url"

	"github.com/net-byte/opensocks/common/util"
	"github.com/net-byte/opensocks/config"
	"github.com/net-byte/opensocks/proxy/httpx"
	"github.com/net-byte/opensocks/proxy/socks"
	p "golang.org/x/net/proxy"
)

var _tcpWorker socks.TCPWorker
var _udpWorker socks.UDPWorker
var _httpServer http.Server

func Start(config config.Config) {
	util.PrintStats(config.Verbose, config.ServerMode)
	if config.HttpProxy {
		// start http proxy server
		go func() {
			socksURL, err := url.Parse("socks5://" + config.LocalAddr)
			if err != nil {
				log.Fatalln("proxy url parse error:", err)
			}
			socks5Dialer, err := p.FromURL(socksURL, p.Direct)
			if err != nil {
				log.Fatalln("failed to make proxy dialer:", err)
			}
			log.Printf("opensocks [http] client started on %s", config.LocalHttpProxyAddr)
			_httpServer = http.Server{
				Addr:    config.LocalHttpProxyAddr,
				Handler: &httpx.HttpProxyHandler{Dialer: socks5Dialer},
			}
			if err := _httpServer.ListenAndServe(); err != nil {
				log.Printf("failed to start http server:%v", err)
			}
		}()
	}
	// start udp worker
	_udpWorker = socks.UDPWorker{Config: config}
	udpConn := _udpWorker.Start()
	// start tcp worker
	_tcpWorker = socks.TCPWorker{Config: config, TCPProxy: &socks.TCPProxy{Config: config}, UDPProxy: &socks.UDPProxy{Config: config}, UDPConn: udpConn}
	_tcpWorker.Start()
}

func Stop() {
	if _tcpWorker.Listener != nil {
		if err := _tcpWorker.Listener.Close(); err != nil {
			log.Printf("failed to shutdown tcp worker: %v", err)
		}
	}
	if _udpWorker.UDPConn != nil {
		if err := _udpWorker.UDPConn.Close(); err != nil {
			log.Printf("failed to shutdown udp worker: %v", err)
		}
	}
	if _httpServer.Addr != "" && _httpServer.Handler != nil {
		if err := _httpServer.Shutdown(context.Background()); err != nil {
			log.Printf("failed to shutdown http server: %v", err)
		}
	}
}
