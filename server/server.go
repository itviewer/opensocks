package server

import (
    "github.com/itviewer/opensocks/common/util"
    "github.com/itviewer/opensocks/server/proxy"
)

var proxyServer proxy.Proxy

func Start() {
    util.PrintStats()
    proxyServer = proxy.NewProxy()
    if proxyServer != nil {
        proxyServer.StartProxyServer()
    }
}

// func Stop() {
//     if proxyServer != nil {
//         proxyServer.StopProxyServer()
//     }
// }
