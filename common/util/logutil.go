package util

import (
    "github.com/itviewer/opensocks/base"
    "time"

    "github.com/itviewer/opensocks/counter"
)

// PrintStats returns the stats info
func PrintStats() {
    if !base.Cfg.Debug {
        return
    }
    go func() {
        // 目前只有本函数使用 CloseChan，所以暂时放在这里
        counter.CloseChan = make(chan struct{})

        ticker := time.NewTicker(10 * time.Second)

        for {
            select {
            case <-ticker.C:
                if base.Cfg.ServerMode {
                    base.Debug("stats:", counter.PrintServerBytes())
                } else {
                    base.Debug("stats:", counter.PrintClientBytes())
                }
            case <-counter.CloseChan:
                ticker.Stop()
                base.Debug("stats timer exit")
                return
            }
        }
    }()
}
