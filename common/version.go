package common

import (
    "github.com/itviewer/opensocks/base"
    "log"
)

var (
    Version   = "v1.8.0"
    GitHash   = ""
    BuildTime = ""
    GoVersion = ""
    // Banner Small Slant
    Banner = `
      ____                ____         __      
     / __ \___  ___ ___  / __/__  ____/ /__ ___
    / /_/ / _ \/ -_) _ \_\ \/ _ \/ __/  '_/(_-<
    \____/ .__/\__/_//_/___/\___/\__/_/\_\/___/
        /_/                                                                                                                 
	`
)

func DisplayVersionInfo() {
    log.Printf("%s", Banner)
    if Version != "" {
        base.Info("version ->", Version)
    }
    if GitHash != "" {
        base.Info("git hash ->", GitHash)
    }
    if BuildTime != "" {
        base.Info("build time ->", BuildTime)
    }
    if GoVersion != "" {
        base.Info("go version ->", GoVersion)
    }
}
