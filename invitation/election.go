package invitation

import (
	"invitation/utils"
	"io"
)

/*
Runs the election, sends only Invites and recieves only Invites
Anything else is discarded
*/
func (st *Status) runElection(at io.ReadWriter) (nextStage uint, err error) {
	backoff := utils.BackoffFrom(int(st.id))
	channel := utils.NewChannelReader(at)
	missing := utils.NewChooser(st.peers)
	nextStage = Member
	for err == nil && (st.leaderId == st.id || len(st.members) != len(st.peers)) {
		stream, readed := backoff.BackoffOnFailure(channel)
		if readed {
			st.checkInvitation(stream, at)
		} else {
			err = st.invitePeer(missing, at)
		}
	}

	if st.leaderId == st.id && err == nil {
		nextStage = Coordinator
	}

	return
}

func (st *Status) checkInvitation(stream []byte, at io.ReadWriter) error {
	if stream[0] == Invite {
		inv, err := deserializeInv(stream[1:])
		if err != nil {
			return err
		}
		if inv.GroupSize > uint(len(st.members)) {
			st.leaderId = inv.Id
			_, err := writeToWithRetry(accept{
				GroupSize: uint(len(st.members)),
				Members:   st.members,
			}, at, st.getPeer(inv.Id))
			return err
		}
	}

	return nil
}

func (st Status) getPeer(peerId uint) string {
	return ""
}

func (st *Status) invitationResponse(response []byte) error {
	//Checks the response to an invitation message
	switch response[0] {
	case Accept:
		//Add the and its group to my members list
	case Reject:
		//Check the id, if 0, the same as the last peer or different
	default:
		//mark the client as dead
	}
	return nil
}

func (st *Status) invitePeer(choser *utils.Choser, at io.ReadWriter) error {
	//Send invitation to peer, rejecting every other invitation with id = 0
	//If peer rejects, and the id is diferent from 0, then send accept and change the leader id
	peer := choser.Choose()
	inv := invite{
		Id:        st.id,
		GroupSize: uint(len(st.members)),
	}
	response, err := writeToWithRetry(inv, at, st.getPeer(peer))

	if err == nil {
		err = st.invitationResponse(response)
	}

	return err
}
