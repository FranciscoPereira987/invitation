package invitation

import (
	"invitation/utils"
	"time"

	"github.com/sirupsen/logrus"
)

func (st *Status) ActAsLeader() (uint, error) {
	st.dial.SetReadDeadline(time.Now().Add(time.Hour * 24))
	msg, addr, err := utils.SafeReadFrom(st.dial)

	if err != nil {
		return Coordinator, err
	}
	switch msg[0] {
	case Invite:
		logrus.Infof("action: acting leader | status: recieved invitation")
		return Electing, st.checkInvitation(msg, addr, nil)
	case Heartbeat:
		//logrus.Infof("action: acting leader | status: recieved heartbeat")
		return Coordinator, writeTo(ok{}, st.dial, addr.String())
	}
	return Coordinator, err
}
