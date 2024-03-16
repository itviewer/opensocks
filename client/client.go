package client

import (
    "context"
    "github.com/itviewer/opensocks/client/proxy/httpx"
    "github.com/itviewer/opensocks/client/proxy/socks5"
    "github.com/itviewer/opensocks/common/util"
    "github.com/itviewer/opensocks/config"
    "golang.org/x/net/proxy"
    "log"
    "net/http"
)

var tcpServer socks5.TCPServer
var udpServer socks5.UDPServer
var httpServer http.Server

func Start(config config.Config) {
    util.PrintStats(config.Verbose, config.ServerMode)
    if config.HttpProxy {
        // start http proxy server
        go func() {
            // http 代理服务器使用本地 socks5 代理和目标服务器建立连接，此时 http 代理服务器作为使用 socks5 代理的普通应用
            socks5Dialer, err := proxy.SOCKS5("tcp", config.LocalAddr, nil, proxy.Direct)
            if err != nil {
                log.Fatalln("failed to make proxy dialer:", err)
            }
            log.Printf("opensocks [http] local server started on %s", config.LocalHttpProxyAddr)
            httpServer = http.Server{
                Addr:    config.LocalHttpProxyAddr,
                Handler: &httpx.HttpProxyHandler{Dialer: socks5Dialer},
            }
            if err := httpServer.ListenAndServe(); err != nil {
                log.Printf("failed to start http server:%v", err)
            }
        }()
    }
    tcpProxy := &socks5.TCPProxy{Config: config}
    udpProxy := &socks5.UDPProxy{Config: config}
    // start udp server
    udpServer = socks5.UDPServer{Config: config, TCPProxy: tcpProxy}
    udpConn := udpServer.Start()
    // start tcp server
    tcpServer = socks5.TCPServer{Config: config, TCPProxy: tcpProxy, UDPProxy: udpProxy, UDPConn: udpConn}
    tcpServer.Start()
}

func Stop() {
    if tcpServer.Listener != nil {
        if err := tcpServer.Listener.Close(); err != nil {
            log.Printf("failed to shutdown tcp worker: %v", err)
        }
    }
    if udpServer.UDPConn != nil {
        if err := udpServer.UDPConn.Close(); err != nil {
            log.Printf("failed to shutdown udp worker: %v", err)
        }
    }
    if httpServer.Addr != "" && httpServer.Handler != nil {
        if err := httpServer.Shutdown(context.Background()); err != nil {
            log.Printf("failed to shutdown http server: %v", err)
        }
    }
}
