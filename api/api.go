package api

import (
    "encoding/json"
    "fmt"
    "github.com/inhies/go-bytesize"
    "github.com/itviewer/opensocks/base"
    "github.com/itviewer/opensocks/client"
    "github.com/itviewer/opensocks/counter"
    "sync"
    "sync/atomic"
)

var closeOnce sync.Once

// StartClient starts the app by json config
func StartClient(jsonConfig string) {
    err := json.Unmarshal([]byte(jsonConfig), &base.Cfg)
    if err != nil {
        fmt.Println("failed to unmarshal config")
        return
    }
    // post init
    base.InitConfig()
    // need log config
    base.InitLog()

    closeOnce = sync.Once{}
    counter.Clean()

    // must not server mode
    // 进入事件循环
    client.Start()
}

// StopClient stops the client
func StopClient() {
    closeOnce.Do(func() {
        client.Stop()
        counter.Clean()
    })
}

// GetTotalReadBytes returns the total read bytes
func GetTotalReadBytes() string {
    return bytesize.New(float64(atomic.LoadUint64(&counter.TotalReadBytes))).String()
}

// GetTotalWrittenBytes returns the total written bytes
func GetTotalWrittenBytes() string {
    return bytesize.New(float64(atomic.LoadUint64(&counter.TotalWrittenBytes))).String()
}
