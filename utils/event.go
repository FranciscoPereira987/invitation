package utils

import (
	"math"
	"math/rand"
	"net"
	"time"
)

const (
	InitialTimeOut      = 1
	TimeoutFactor       = 2
	NanoToSecondsFactor = 9
)

type BackoffTimer struct {
	maxTimeout float64
	source     *rand.Rand
}

func BackoffFrom(seed int) *BackoffTimer {
	source := rand.NewSource(int64(seed))

	return &BackoffTimer{
		maxTimeout: InitialTimeOut,
		source:     rand.New(source),
	}
}

func (bckoff *BackoffTimer) IncreaseTimeOut() {
	bckoff.maxTimeout *= TimeoutFactor
}

func (bckoff *BackoffTimer) TimeOut() time.Time {
	timeOut := bckoff.source.Int63n(int64(bckoff.maxTimeout * math.Pow10(NanoToSecondsFactor)))
	return time.Now().Add(time.Duration(timeOut))
}

func (bckoff *BackoffTimer) SetReadTimeout(sckt *net.UDPConn) {
	sckt.SetReadDeadline(bckoff.TimeOut())
}

func (bckoff *BackoffTimer) Wait() {
	sleepTime := bckoff.source.Int63n(int64(bckoff.maxTimeout * math.Pow10(NanoToSecondsFactor)))
	time.Sleep(time.Duration(sleepTime))
}
