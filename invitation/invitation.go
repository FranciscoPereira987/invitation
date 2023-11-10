package invitation

const (
	Electing = iota
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

type Status struct {
	peers []uint
	members []uint

	id uint
	leaderId uint
}

func Invitation(peers []uint, id uint) *Status {
	return &Status{
		peers: peers,
		members: make([]uint, 0),
		id: id,
		leaderId: id,
	}
}

func (st *Status) Run() error {
	state := Electing
	for {
		switch state{
		case Electing:
			//Running an election
		case Coordinator:
			//Leader
		case Member:
			//Member of a group
		}
	}

	return nil
}