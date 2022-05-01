package connector

import (
	"math"
	"sync/atomic"
)

var (
	maxClients int32 = math.MaxInt32
	numClients int32
)

func SetMaxClients(v int32) {
	atomic.StoreInt32(&maxClients, v)
}

func MaxClients() int {
	return int(atomic.LoadInt32(&maxClients))
}

func NumClients() int {
	return int(atomic.LoadInt32(&numClients))
}

func ExceedMaxClients() bool {
	return NumClients() >= MaxClients()
}

func incrNumClients() {
	atomic.AddInt32(&numClients, 1)
}

func decrNumClients() {
	atomic.AddInt32(&numClients, -1)
}
