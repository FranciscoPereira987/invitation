package utils

import (
	"math"
	"math/rand"
	"time"
)

const (
	InitialTimeOut = 1
	TimeoutFactor = 2
	NanoToSecondsFactor = 9
)

type BackoffTimer struct {
	maxTimeout float64
	source *rand.Rand 
}

func BackoffFrom(seed int) *BackoffTimer {
	source := rand.NewSource(int64(seed))

	return &BackoffTimer{
		maxTimeout: InitialTimeOut,
		source: rand.New(source),
	}
}

func (bckoff *BackoffTimer) IncreaseTimeOut() {
	bckoff.maxTimeout *= TimeoutFactor
}

func (bckoff *BackoffTimer) TimeOut() <- chan time.Time {
	timeOut := bckoff.source.Int63n(int64(bckoff.maxTimeout * math.Pow10(NanoToSecondsFactor)))
	return time.After(time.Duration(timeOut))
}

func (bckoff *BackoffTimer) BackoffOnFailure(sckt *ChannelReader) (stream []byte, readed bool) {
	stream, readed = TimerEvent(sckt, bckoff)
	if !readed {
		bckoff.IncreaseTimeOut()
	}
	return
}

/*
	Returns true if the sckt was read before the timer timed out
*/
func TimerEvent(sckt *ChannelReader, bckoff *BackoffTimer) ([]byte, bool) {
	select {
	case stream := <- sckt.Channel:
		return stream, true
	case <- bckoff.TimeOut():
		return nil, false
	}
}