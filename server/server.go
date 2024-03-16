package server

import (
    "github.com/itviewer/opensocks/common/util"
    "github.com/itviewer/opensocks/config"
    "github.com/itviewer/opensocks/server/proxy"
)

var proxyServer proxy.Proxy

func Start(config config.Config) {
    util.PrintStats(config.Verbose, config.ServerMode)
    proxyServer = proxy.NewProxy(config)
    proxyServer.StartServer()
}

func Stop() {
    proxyServer.StopServer()
}
