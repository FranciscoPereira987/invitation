package invitation

import (
	"invitation/utils"
	"net"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	Electing uint = iota
	Coordinator
	Member
)

const (
	Invite = iota + 1
	Reject
	Accept
	Change
	Heartbeat
	Ok
)

type invite struct {
	Id        uint
	GroupSize uint
}

type reject struct {
	LeaderId uint
}

type accept struct {
	GroupSize uint
	Members   []uint
}

type change struct {
	NewLeaderId uint
}

type heartbeat struct{}

type ok struct{}

type Status struct {
	peers *utils.Peers

	dial *net.UDPConn

	id       uint
	leaderId uint
}

func Invitation(peers []uint, id uint, mapping map[uint]string, conn *net.UDPConn) *Status {
	return &Status{
		peers:    utils.NewPeers(peers, mapping),
		id:       id,
		dial:     conn,
		leaderId: id,
	}
}

func (st *Status) Run() (err error) {
	state := Electing
	for err == nil {
		switch state {
		case Electing:
			st.leaderId = st.id
			state, err = st.runElection()
		case Coordinator:
			//Leader
			logrus.Infof("Action: peer %d acting as leader", st.id)
			state, err = st.ActAsLeader()
		case Member:
			//Member of a group
			logrus.Infof("Action: peer %d acting as member | leader: peer %d ", st.id, st.leaderId)
			state, err = st.ActAsMember()
		}
	}

	return
}

func writeTo(s serializable, at *net.UDPConn, where string) error {
	//Writes the whole stream into at, if failed to do so, returns an error
	addr, err := net.ResolveUDPAddr("udp", where)
	if err != nil {
		return err
	}

	return utils.SafeWriteTo(s.serialize(), at, addr)
}

func writeToWithRetry(s serializable, at *net.UDPConn, where string) ([]byte, error) {
	//Tries to pass a stream to the other endpoint and awaits a response
	//It tries three times, else, an error is returned
	//If the response is not from the peer i sent the message, sends a reject 0
	retries := 3
	backoff := utils.BackoffFrom(time.Now().Nanosecond())
	var err error
	var buf []byte
	for ; retries > 0 && err == nil; retries-- {
		err = writeTo(s, at, where)
		if err == nil {
			backoff.SetReadTimeout(at)
			buf, err = readFrom(at, where)
		}
		if err == nil {
			return buf, err
		}
		backoff.IncreaseTimeOut()
	}

	return nil, err
}

func readFrom(from *net.UDPConn, expected string) ([]byte, error) {
	//Tries reading from io.Reader
	//Returns the whole array of bytes readed if successful.
	expectedAddr, err := net.ResolveUDPAddr("udp", expected)
	if err != nil {
		return nil, err
	}
	buf, addr, err := utils.SafeReadFrom(from)
	if err == nil && !addr.IP.Equal(expectedAddr.IP) {
		writeTo(reject{
			0,
		}, from, addr.String())
		return readFrom(from, expected)
	}
	return buf, err
}
