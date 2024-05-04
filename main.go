package main

import (
    "flag"
    "github.com/itviewer/opensocks/base"

    "github.com/itviewer/opensocks/client"
    "github.com/itviewer/opensocks/common"
    "github.com/itviewer/opensocks/server"
)

func main() {
    flag.StringVar(&base.Cfg.LocalAddr, "l", "127.0.0.1:1080", "local socks5 proxy address")
    flag.StringVar(&base.Cfg.LocalHttpProxyAddr, "http", ":8008", "local http proxy address")
    flag.StringVar(&base.Cfg.ServerAddr, "s", ":8081", "server address")
    flag.StringVar(&base.Cfg.Key, "k", "6w9z$C&F)J@NcRfUjXn2r4u7x!A%D*G-", "encryption key")
    flag.BoolVar(&base.Cfg.ServerMode, "S", false, "server mode")
    flag.StringVar(&base.Cfg.Protocol, "p", "tcp", "protocol tcp/ws/kcp")
    flag.BoolVar(&base.Cfg.Bypass, "bypass", false, "bypass private ip")
    flag.BoolVar(&base.Cfg.Obfs, "obfs", false, "enable data obfuscation")
    flag.BoolVar(&base.Cfg.Compress, "compress", false, "enable data compression")
    flag.BoolVar(&base.Cfg.HttpProxy, "http-proxy", false, "enable http proxy")
    flag.BoolVar(&base.Cfg.Debug, "v", false, "enable verbose output")
    flag.Parse()

    base.InitConfig()
    base.InitLog()

    common.DisplayVersionInfo()

    // 进入事件循环
    if base.Cfg.ServerMode {
        server.Start()
    } else {
        client.Start()
    }
}
