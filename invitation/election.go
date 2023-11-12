package invitation

import (
	"invitation/utils"
	"net"
	"time"

	"github.com/sirupsen/logrus"
)

/*
Runs the election, sends only Invites and recieves only Invites
Anything else is discarded
*/
func (st *Status) runElection() (nextStage uint, err error) {
	missing := utils.NewChooser(st.peers.Peers)
	nextStage = Member
	backoff := utils.BackoffFrom(time.Now().Nanosecond())
	for err == nil && st.leaderId == st.id && !st.peers.GroupIsComplete() && missing.PeersLeft() {
		backoff.SetReadTimeout(st.dial)
		logrus.Info("action: election | status: waiting for peer messages")
		stream, addr, err := utils.SafeReadFrom(st.dial)
		logrus.Infof("action: election | status: stream read | result: %s | stream: %s", err, stream)
		if err == nil {
			err = st.checkInvitation(stream, addr)
		} else {
			err = st.invitePeer(missing)
			logrus.Infof("action: election | status: peer invitation | result: %s", err)
		}
	}

	if st.leaderId == st.id && err == nil {
		nextStage = Coordinator
	}

	return
}

func (st *Status) checkInvitation(stream []byte, to *net.UDPAddr) error {
	switch stream[0] {
	case Invite:
		inv, err := deserializeInv(stream[1:])
		if err != nil {
			return err
		}
		if inv.GroupSize > uint(len(st.peers.Members)) {
			err = st.accept(inv.Id)
			return err
		} else {
			err = st.reject(inv.Id)
			return err
		}
	case Accept:
		err := st.invitationResponse(stream, 0)
		return err
	case Heartbeat:
		return writeTo(ok{}, st.dial, to.String())
	}

	return nil
}

func (st Status) getPeer(peerId uint) string {
	return st.peers.GetAddr(peerId)
}

func (st *Status) addToGroup(peer uint, peerGroup []uint) {
	st.peers.AddMembers(append(peerGroup, peer)...)
}

/*
After having sent an invitation to some peer and the peer has answered me I need to:

 1. If the response was an accept, then add the peer and its group to my group

 2. If the response was a reject:
    i- With LeaderID 0, then the peer could not give me a proper answer
    ii- With LeaderID equal that of the the peer, then that peer is my new leader
    iii- With LeaderID different from that of the peer, then I should invite its leader.

 3. Anything else, I should mark the peer as dead
*/
func (st *Status) invitationResponse(response []byte, lastPeer uint) error {
	switch response[0] {
	case Accept:
		acc, err := deserializeAcc(response[1:])
		if err != nil {
			return err
		}
		st.addToGroup(lastPeer, acc.Members)
	case Reject:
		rej, err := deserializeRej(response[1:])
		if err != nil {
			return err
		}
		if rej.LeaderId != 0 {
			if rej.LeaderId == lastPeer {
				st.accept(lastPeer)
			} else {
				st.invite(rej.LeaderId)
			}
		}
	}
	return nil
}

func (st *Status) invite(peer uint) error {
	inv := invite{
		Id:        st.id,
		GroupSize: uint(len(st.peers.Members)),
	}
	logrus.Infof("Action: election | status: Inviting peer %d", peer)
	response, err := writeToWithRetry(inv, st.dial, st.getPeer(peer))

	if err == nil {
		logrus.Infof("action: election | status: Response from peer %d", peer)
		err = st.invitationResponse(response, peer)
	}

	return err
}

func (st *Status) reject(peer uint) error {
	rej := reject{
		LeaderId: st.leaderId,
	}
	return writeTo(rej, st.dial, st.getPeer(peer))
}

/*
Accepts the invitation from a peer and deletes its own group
*/
func (st *Status) accept(peer uint) error {
	st.leaderId = peer
	acc := accept{
		GroupSize: uint(len(st.peers.Members)),
		Members:   st.peers.Members,
	}
	if err := writeTo(acc, st.dial, st.getPeer(peer)); err != nil {
		return err
	}
	st.change()
	st.peers.Members = make([]uint, 0)
	return nil
}

/*
Sends change message to all members of the group
*/
func (st *Status) change() error {
	ch := change{
		NewLeaderId: st.leaderId,
	}

	for _, member := range st.peers.Members {
		if err := writeTo(ch, st.dial, st.getPeer(member)); err != nil {
			return err
		}
	}

	return nil
}

func (st *Status) invitePeer(choser *utils.Choser) error {
	//Send invitation to peer, rejecting every other invitation with id = 0
	//If peer rejects, and the id is diferent from 0, then send accept and change the leader id
	peer := choser.Choose()
	for ; st.peers.IsMember(peer); peer = choser.Choose() {
	}
	return st.invite(peer)
}
