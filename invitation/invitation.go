package invitation

import (
	"io"
)

const (
	Electing uint = iota
	Coordinator
	Member
)

const (
	Invite = iota
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
	peers   []uint
	members []uint

	id       uint
	leaderId uint
}

func Invitation(peers []uint, id uint) *Status {
	return &Status{
		peers:    peers,
		members:  make([]uint, 0),
		id:       id,
		leaderId: id,
	}
}

func (st *Status) Run(at io.ReadWriter) (err error) {
	state := Electing
	for err == nil {
		switch state {
		case Electing:
			state, err = st.runElection(at)
		case Coordinator:
			//Leader
		case Member:
			//Member of a group
		}
	}

	return
}

func writeTo(s serializable, at io.Writer, where string) error {
	//Writes the whole stream into at, if failed to do so, returns an error
	return nil
}

func writeToWithRetry(s serializable, at io.ReadWriter, where string) ([]byte, error) {
	//Tries to pass a stream to the other endpoint and awaits a response
	//It tries three times, else, an error is returned
	//If the response is not from the peer i sent the message, sends a reject 0
	return nil, nil
}

func readFrom(from io.Reader) ([]byte, error) {
	//Tries reading from io.Reader
	//Returns the whole array of bytes readed if successful.
	return nil, nil
}
