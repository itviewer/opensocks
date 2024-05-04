package main

import "C"
import (
    "github.com/itviewer/opensocks/api"
)

func main() {}

//export apiStartClient
func apiStartClient(jsonConfig string) {
    api.StartClient(jsonConfig)
}

//export apiStopClient
func apiStopClient() {
    api.StopClient()
}

//export apiGetDownloadByteSize
func apiGetDownloadByteSize() *C.char {
    return C.CString(api.GetTotalReadBytes())
}

//export apiGetUploadByteSize
func apiGetUploadByteSize() *C.char {
    return C.CString(api.GetTotalWrittenBytes())
}
