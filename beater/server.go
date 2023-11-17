package beater

import (
	"invitation/utils"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	Heartbeat = iota + 1
	Ok
)

type heartbeat struct{}

func (hb heartbeat) serialize() []byte {
	return []byte{Heartbeat}
}

type ok struct {
	whom string
}

type client struct {
	Name string
	Addr string
}

func (o ok) serialize() []byte {
	header := []byte{Ok}
	body := utils.EncodeString(o.whom)
	return append(header, body...)
}

func (o *ok) deserialize(stream []byte) error {
	decoded, err := utils.DecodeString(stream)
	if err == nil {
		o.whom = decoded
	}
	return err
}

type url = any

/*
The Beater Server makes sure that all uniquely named clients
are alive.
*/
type BeaterServer struct {
	clients map[string]*timer

	clientInfo []client

	sckt *net.UDPConn

	wg *sync.WaitGroup
}

/*
Timer runs a routine to keep clients
in check, ensuring that they are alive
*/
type timer struct {
	//Allows for the timer to reset
	//if Inbound is false, then the routine shutsdown
	InboundChan chan bool
	//Outputs a message to be delivered to the client
	OutboundChan chan *net.UDPAddr

	clientAddr *net.UDPAddr

	/*
		Service name that's going to be used to
		recreate the client
	*/
	name string

	//Heartbeat timer
	maxTime time.Duration
}

func NewTimer(at string, name string) (*timer, error) {
	addr, err := net.ResolveUDPAddr("udp", at)
	t := new(timer)
	if err == nil {
		t.InboundChan = make(chan bool, 1)
		t.OutboundChan = make(chan *net.UDPAddr, 1)
		t.clientAddr = addr
		t.name = name
		t.maxTime = time.Second * 2
	}
	return t, err
}

func (t *timer) executeTimer(group *sync.WaitGroup) {
	defer close(t.OutboundChan)
loop:
	for {
		timeout := time.After(t.maxTime)
		t.OutboundChan <- t.clientAddr
		select {
		case <-timeout:
			logrus.Info("Service is dead")
		case result := <-t.InboundChan:
			logrus.Info("Client answered")
			if !result {
				logrus.Infof("action: Client %s timer | status: ending", t.clientAddr.String())
				break loop
			}
			<-time.After(t.maxTime / 2)
		}
	}
	group.Done()

}

func NewBeaterServer(clients []string, clientAddrs []string, at string) *BeaterServer {
	client_map := make(map[string]*timer)

	clientInfo := make([]client, 0)
	for index, clientName := range clients {
		clientInfo = append(clientInfo, client{
			clientName,
			clientAddrs[index],
		})
	}

	addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:"+at)
	if err != nil {
		return nil
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil
	}
	return &BeaterServer{
		client_map,
		clientInfo,
		conn,
		new(sync.WaitGroup),
	}
}

/*
Initiates timers and merge channels
*/
func (b *BeaterServer) initiateTimers(port string) chan *net.UDPAddr {
	mergedChans := make([](<-chan *net.UDPAddr), 0)
	for _, value := range b.clientInfo {

		timer, err := NewTimer(value.Addr+":"+port, value.Name)
		if err == nil {
			b.clients[value.Name] = timer
			mergedChans = append(mergedChans, timer.OutboundChan)
			go timer.executeTimer(b.wg)
		} else {
			logrus.Fatalf("action: timer-initialization | status: failed | reason: %s", err)
		}
	}
	b.wg.Add(len(mergedChans))
	return utils.Merge(mergedChans...)
}

/*
Initiates socket reading routine
*/
func (b *BeaterServer) initiateReader() chan []byte {
	readerChan := make(chan []byte, 1)
	go func() {
		var err error
		for err == nil {
			beat, _, err_read := utils.SafeReadFrom(b.sckt)
			err = err_read
			if err == nil {
				readerChan <- beat
			}
		}
	}()
	return readerChan
}

func (b *BeaterServer) parseBeat(beat []byte) {
	if beat[0] != Ok {
		logrus.Errorf("recieved invalid beat response")
		return
	}
	ok := &ok{}
	err := ok.deserialize(beat[1:])
	if err != nil {
		logrus.Errorf("error while deserializing beat response: %s", err)
		return
	}
	if timer, ok := b.clients[ok.whom]; ok {
		timer.InboundChan <- true
	}
}

/*
Runs a routine that is in charge of writing to the socket
heartbeats
*/
func (b *BeaterServer) writeRoutine(channel <-chan *net.UDPAddr) {
	for addr := range channel {
		err := utils.SafeWriteTo(heartbeat{}.serialize(), b.sckt, addr)
		if err != nil {
			logrus.Errorf("action: heartbeat to services | status: %s", err)
		}
	}
	logrus.Info("action: write routine | status: finishing")
}

/*
1. server starts all timers
2. server loop:
  - Select -> write to client channel
    -> Send to write rutine
    -> listen from socket channel
    -> Update client mappings
    -> listen for shutdown
    -> shutdown if necessary
*/
func (b *BeaterServer) Run(port string) (err error) {
	timersChan := b.initiateTimers(port)
	defer close(timersChan)
	readerChan := b.initiateReader()
	defer close(readerChan)
	go b.writeRoutine(timersChan)

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGTERM)

loop:
	for {
		select {
		case beat := <-readerChan:
			b.parseBeat(beat)
		case <-shutdown:
			break loop
		}
	}

	for _, t := range b.clients {
		t.InboundChan <- false
	}

	b.wg.Wait()

	return
}
