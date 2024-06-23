package server

import (
    "github.com/itviewer/opensocks/base"
    "github.com/itviewer/opensocks/common/util"
    "github.com/itviewer/opensocks/server/proxy"
    "os"
    "os/signal"
    "syscall"
)

var (
    tcpProxyServer *proxy.TCPProxy
    kcpProxyServer *proxy.KCPProxy
)

func Start() {
    util.PrintStats()

    tcpProxyServer = &proxy.TCPProxy{}
    kcpProxyServer = &proxy.KCPProxy{}

    go tcpProxyServer.StartProxyServer()
    go kcpProxyServer.StartProxyServer()

    watchSignal()
}

func Stop() {
    if tcpProxyServer != nil {
        tcpProxyServer.StopProxyServer()
    }
    if kcpProxyServer != nil {
        kcpProxyServer.StopProxyServer()
    }
}

func watchSignal() {
    base.Info("Server pid: ", os.Getpid())

    sigs := make(chan os.Signal, 1)
    // https://pkg.go.dev/os/signal
    signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
    for {
        // 没有信号就阻塞，从而避免主协程退出
        sig := <-sigs
        base.Info("Get signal:", sig)
        switch sig {
        default:
            base.Info("Stop")
            Stop()
            return
        }
    }
}
