package beater

import (
	"invitation/utils"
	"net"

	"github.com/sirupsen/logrus"
)

/*
Each client has a name which serves as an ID of the client
*/
type BeaterClient struct {
	conn *net.UDPConn

	name string
}

func NewBeaterClient(name string, addr string) (*BeaterClient, error) {
	address, err := net.ResolveUDPAddr("udp", addr)

	if err != nil {
		return nil, err
	}

	conn, err := net.ListenUDP("udp", address)

	return &BeaterClient{
		conn,
		name,
	}, err
}

func (st *BeaterClient) Run() error {
	var err error
	for err == nil {
		recovered, server, err_read := utils.SafeReadFrom(st.conn)
		logrus.Info("action: client | status: recieved heartbeat from server")
		err = err_read
		if err == nil {
			if recovered[0] == Heartbeat {
				logrus.Info("action: client | status: responding to server")
				err = utils.SafeWriteTo(ok{st.name}.serialize(), st.conn, server)
			}
		}
	}
	return err
}
