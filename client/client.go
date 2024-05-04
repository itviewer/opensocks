package client

import (
    "context"
    "github.com/itviewer/opensocks/base"
    "github.com/itviewer/opensocks/client/proxy/httpx"
    "github.com/itviewer/opensocks/client/proxy/socks5"
    "github.com/itviewer/opensocks/common/util"
    "github.com/itviewer/opensocks/counter"
    "golang.org/x/net/proxy"
    "net/http"
)

var (
    tcpServer  socks5.TCPServer
    udpServer  socks5.UDPServer
    httpServer http.Server
)

func Start() {
    util.PrintStats()

    if base.Cfg.HttpProxy {
        // start http proxy server
        go func() {
            // http 代理服务器使用本地 socks5 代理和目标服务器建立连接，此时 http 代理服务器作为使用 socks5 代理的普通应用
            socks5Dialer, err := proxy.SOCKS5("tcp", base.Cfg.LocalAddr, nil, proxy.Direct)
            if err != nil {
                base.Error("failed to make proxy dialer:", err)
            }
            base.Info("http server started on", base.Cfg.LocalHttpProxyAddr)
            httpServer = http.Server{
                Addr:    base.Cfg.LocalHttpProxyAddr,
                Handler: &httpx.HttpProxyHandler{Dialer: socks5Dialer},
            }
            if err := httpServer.ListenAndServe(); err != nil {
                base.Error("http server:", err)
            }
        }()
    }
    tcpProxy := &socks5.TCPProxy{}
    udpProxy := &socks5.UDPProxy{}

    // start udp server
    udpServer = socks5.UDPServer{TCPProxy: tcpProxy}
    udpConn := udpServer.Start()

    // start tcp server
    tcpServer = socks5.TCPServer{TCPProxy: tcpProxy, UDPProxy: udpProxy, UDPConn: udpConn}
    tcpServer.Start() // 进入事件循环

    // 确保退出事件循环关闭打印定时器
    if base.Cfg.Debug {
        close(counter.CloseChan)
    }
}

func Stop() {
    if udpServer.UDPConn != nil {
        if err := udpServer.UDPConn.Close(); err != nil {
            base.Error("failed to shutdown udp local server:", err)
        }
    }

    if httpServer.Addr != "" && httpServer.Handler != nil {
        if err := httpServer.Shutdown(context.Background()); err != nil {
            base.Error("failed to shutdown http local server:", err)
        }
    }

    if tcpServer.Listener != nil {
        if err := tcpServer.Listener.Close(); err != nil {
            base.Error("failed to shutdown tcp local server:", err)
        }
    }
}
