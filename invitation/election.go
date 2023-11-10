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
			st.checkInvitation(stream)
		}else{
			err = st.invitePeer(missing, at)
		}
	}
	
	if st.leaderId == st.id && err == nil {
		nextStage = Coordinator
	}

	return
}

func (st *Status) checkInvitation(stream []byte) {
	if stream[0] == Invite {
		//Parse invitation
		//Check if group size is greater than mine
	}
}

func (st *Status) invitePeer(choser *utils.Choser, at io.ReadWriter) error {
	//Send invitation to peer, rejecting every other invitation with id = 0
	//If peer rejects, and the id is diferent from 0, then send accept and change the leader id
	return nil
}
