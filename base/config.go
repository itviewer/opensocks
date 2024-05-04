package base

import (
    "github.com/itviewer/opensocks/common/cipher"
)

var (
    Cfg = &Config{LogLevel: "Info"}
)

type Config struct {
    LocalAddr          string
    LocalHttpProxyAddr string
    ServerAddr         string
    Key                string
    Protocol           string
    ServerMode         bool
    Bypass             bool
    Obfs               bool
    Compress           bool
    HttpProxy          bool
    Debug              bool

    LogPath  string
    LogLevel string
}

// 无论是命令行还是 api，都会传默认值，这里非必须
func init() {
    // Cfg.LocalAddr = "127.0.0.1:1080"
    // Cfg.LocalHttpProxyAddr = ":8008"
    // Cfg.ServerAddr = ":8081"
    // Cfg.Key = "6w9z$C&F)J@NcRfUjXn2r4u7x!A%D*G-"
    // Cfg.Protocol = "tcp"
    // Cfg.ServerMode = false
    // Cfg.Bypass = false
    // Cfg.Obfs = false
    // Cfg.Compress = false
    // Cfg.HttpProxy = false
    // Cfg.Debug = false

    // Cfg.LogLevel = "Info"
}

// InitConfig post init
func InitConfig() {
    if Cfg.Debug {
        Cfg.LogLevel = "Debug"
    }
    cipher.GenerateKey(Cfg.Key)
}
