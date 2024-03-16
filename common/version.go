package common

import "log"

var (
    Version   = "v1.8.0"
    GitHash   = ""
    BuildTime = ""
    GoVersion = ""
    Banner    = `
	___                                        _        
	/ _ \   _ __   ___   _ _    ___  ___   __  | |__  ___
	| (_) | | '_ \ / -_) | ' \  (_-< / _ \ / _| | / / (_-<
	\___/  | .__/ \___| |_||_| /__/ \___/ \__| |_\_\ /__/
		 |_|                                           
	`
)

func DisplayVersionInfo() {
    log.Printf("%s", Banner)
    if Version != "" {
        log.Printf("version -> %s", Version)
    }
    if GitHash != "" {
        log.Printf("git hash -> %s", GitHash)
    }
    if BuildTime != "" {
        log.Printf("build time -> %s", BuildTime)
    }
    if GoVersion != "" {
        log.Printf("go version -> %s", GoVersion)
    }
}
