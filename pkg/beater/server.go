package beater

import (
	"errors"
	"invitation/pkg/utils"
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

type Runable interface {
	Stop() error
	Run()
}	

/*
The Beater Server makes sure that all uniquely named clients
are alive.
*/
type BeaterServer struct {
	clients map[string]*timer

	clientInfo []client

	sckt *net.UDPConn

	wg *sync.WaitGroup

	resultsChan chan error

	port string

	shutdown chan os.Signal
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

	clientAddr string

	/*
		Service name that's going to be used to
		recreate the client
	*/
	name string

	//Heartbeat timer
	maxTime time.Duration
}

func NewTimer(at string, name string) *timer {
	t := new(timer)
	
	t.InboundChan = make(chan bool, 1)
	t.OutboundChan = make(chan *net.UDPAddr, 1)
	t.clientAddr = at
	t.name = name
	t.maxTime = time.Second * 2
	
	return t
}

func (t *timer) executeTimer(group *sync.WaitGroup) {
	defer close(t.OutboundChan)
	clientAddr, err := net.ResolveUDPAddr("udp", t.clientAddr)
	for err != nil {
		logrus.Errorf("action: resolving client address | result: failed | action: re-instantiating client")
		<- time.After(t.maxTime * 5)
		clientAddr, err = net.ResolveUDPAddr("udp", t.clientAddr)
		select {
		case <- t.InboundChan:
			group.Done()
			return
		default:
		}
	}
loop:
	for {
		timeout := time.After(t.maxTime)
		t.OutboundChan <- clientAddr
		select {
		case <-timeout:
			logrus.Info("Service is dead")
		case result := <-t.InboundChan:
			logrus.Info("Client answered")
			if !result {
				logrus.Infof("action: Client %s timer | status: ending", t.clientAddr)
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
		make(chan error, 1),
		at,
		nil,
	}
}

/*
Initiates timers and merge channels
*/
func (b *BeaterServer) initiateTimers(port string) chan *net.UDPAddr {
	mergedChans := make([](<-chan *net.UDPAddr), 0)
	for _, value := range b.clientInfo {

		timer := NewTimer(value.Addr+":"+port, value.Name)
		
		b.clients[value.Name] = timer
		mergedChans = append(mergedChans, timer.OutboundChan)
		go timer.executeTimer(b.wg)
		
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

func (b *BeaterServer) run(port string) (err error) {
	timersChan := b.initiateTimers(port)
	defer close(timersChan)
	readerChan := b.initiateReader()
	defer close(readerChan)
	go b.writeRoutine(timersChan)

	shutdown := make(chan os.Signal, 1)
	b.shutdown = shutdown
	defer close(shutdown)
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
func (b *BeaterServer) Run() {
	go func() {
		b.resultsChan <- b.run(b.port)
	}()
}

func (b *BeaterServer) Stop() error {
	err := b.sckt.Close()
	b.shutdown <- syscall.SIGTERM
	err = errors.Join(err, <- b.resultsChan)
	
	return err
}
