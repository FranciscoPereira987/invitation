package invitation

import (
	"invitation/pkg/utils"
	"time"

	"github.com/sirupsen/logrus"
)

func (st *Status) checkMissing() {
	missing := st.peers.GetMissing()
	if missing != nil {
		for _, peer := range missing {
			writeTo(invite{
				Id:        st.id,
				GroupSize: uint(len(st.peers.Members)),
			}, st.dial, st.getPeer(peer))
		}
	}
}

func (st *Status) ActAsLeader() (uint, error) {
	st.dial.SetReadDeadline(time.Now().Add(time.Hour * 24))
	msg, addr, err := utils.SafeReadFrom(st.dial)

	if err != nil {
		return Coordinator, err
	}
	st.checkMissing()
	switch msg[0] {
	case Invite:
		logrus.Infof("action: acting leader | status: recieved invitation")
		st.checkInvitation(msg, addr, nil)
		return st.runElection()
	case Heartbeat:
		//logrus.Infof("action: acting leader | status: recieved heartbeat")
		return Coordinator, writeTo(ok{}, st.dial, addr.String())
	}
	return Coordinator, err
}
