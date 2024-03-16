package pool

import (
    "github.com/itviewer/opensocks/common/enum"
    "sync"
)

var pool = sync.Pool{
    New: func() interface{} {
        b := make([]byte, enum.BufferSize)
        return b
    },
}

func Get() []byte {
    buf := pool.Get().([]byte)
    return buf
}

func Put(buf []byte) {
    pool.Put(buf)
}
