package server

import (
	"github.com/net-byte/opensocks/common/util"
	"github.com/net-byte/opensocks/config"
	"github.com/net-byte/opensocks/proxy"
)

var _proxyServer proxy.Proxy

func Start(config config.Config) {
	util.PrintStats(config.Verbose, config.ServerMode)
	_proxyServer = proxy.NewProxy(config)
	_proxyServer.StartServer()
}

func Stop() {
	_proxyServer.StopServer()
}
