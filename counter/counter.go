package counter

import (
	"fmt"
	"sync/atomic"

	"github.com/inhies/go-bytesize"
)

// TotalReadBytes is the total number of bytes read
var TotalReadBytes uint64 = 0

// TotalWrittenBytes is the total number of bytes written
var TotalWrittenBytes uint64 = 0

// IncrReadBytes increments the number of bytes read
func IncrReadBytes(n int) {
	atomic.AddUint64(&TotalReadBytes, uint64(n))
}

// IncrWrittenBytes increments the number of bytes written
func IncrWrittenBytes(n int) {
	atomic.AddUint64(&TotalWrittenBytes, uint64(n))
}

// PrintClientBytes returns the bytes info on client side
func PrintClientBytes() string {
	return fmt.Sprintf("download %v upload %v", bytesize.New(float64(atomic.LoadUint64(&TotalReadBytes))).String(), bytesize.New(float64(atomic.LoadUint64(&TotalWrittenBytes))).String())
}

// PrintServerBytes returns the bytes info on server side
func PrintServerBytes() string {
	return fmt.Sprintf("download %v upload %v", bytesize.New(float64(atomic.LoadUint64(&TotalWrittenBytes))).String(), bytesize.New(float64(atomic.LoadUint64(&TotalReadBytes))).String())
}

// Clean clean the counter
func Clean() {
	TotalReadBytes = 0
	TotalWrittenBytes = 0
}
