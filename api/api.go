package api

import (
    "encoding/json"
    "log"
    "strconv"
    "sync/atomic"

    "github.com/itviewer/opensocks/client"
    "github.com/itviewer/opensocks/config"
    "github.com/itviewer/opensocks/counter"
    "github.com/itviewer/opensocks/server"
)

// Start starts the app by json config
func Start(jsonConfig string) {
    CleanCounter()
    config := config.Config{}
    err := json.Unmarshal([]byte(jsonConfig), &config)
    if err != nil {
        log.Panic("failed to decode config")
    }
    config.Init()
    if config.ServerMode {
        server.Start(config)
    } else {
        client.Start(config)
    }
}

// StopClient stops the client
func StopClient() {
    client.Stop()
    CleanCounter()
}

// StopServer stops the server
func StopServer() {
    server.Stop()
    CleanCounter()
}

// GetTotalReadBytes returns the total read bytes
func GetTotalReadBytes() string {
    return strconv.FormatUint(atomic.LoadUint64(&counter.TotalReadBytes), 10)
}

// GetTotalWrittenBytes returns the total written bytes
func GetTotalWrittenBytes() string {
    return strconv.FormatUint(atomic.LoadUint64(&counter.TotalWrittenBytes), 10)
}

// CleanCounter cleans the counter
func CleanCounter() {
    counter.Clean()
}
