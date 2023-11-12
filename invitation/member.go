package invitation

import (
	"invitation/utils"
	"time"

	"github.com/sirupsen/logrus"
)

func (st *Status) ActAsMember() (uint, error) {
	timer := utils.BackoffFrom(time.Now().Nanosecond())
	timer.SetReadTimeout(st.dial)
	msg, addr, err := utils.SafeReadFrom(st.dial)
	if err == nil {
		switch msg[0] {
		case Invite:
			logrus.Infof("action: acting as member | status: recieved invite from: %s", addr.String())
			err = writeTo(reject{
				LeaderId: st.leaderId,
			}, st.dial, addr.String())
		case Change:
			ch, err := deserializeChange(msg[1:])
			if err == nil {
				logrus.Infof("action: acting as member | status: recieved change from leader to: %d", ch.NewLeaderId)
				st.leaderId = ch.NewLeaderId
			}
		}
	}
	_, err = writeToWithRetry(heartbeat{}, st.dial, st.getPeer(st.leaderId))
	if err != nil {
		return Electing, nil
	}
	return Member, err
}
