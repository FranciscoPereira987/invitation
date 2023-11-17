package main

import (
	"errors"
	"invitation/beater"
	"invitation/invitation"
	"invitation/utils"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const (
	Name        = "name"
	NetPort     = "net_port"
	Peers       = "peers"
	PeerName    = "peer_name"
	PeerNetName = "net_name"
)

var (
	InvalidPeerNameErr    = errors.New("invalid peer name")
	InvalidPeerNetNameErr = errors.New("invalid peer net name")
)

func parseConfig(v *viper.Viper) (config *invitation.Config, err error) {
	port := v.GetString(NetPort)
	id := v.GetInt32(Name)

	config, err = invitation.NewConfigOn(uint(id), port)
	if err == nil {
		mapped := v.Get(Peers).([]interface{})
		for _, value := range mapped {
			value := value.(map[string]any)
			peerName, ok := value[PeerName].(int)
			if !ok {
				err = InvalidPeerNameErr
				return
			}
			peerNetName, ok := value[PeerNetName].(string)
			if !ok {
				err = InvalidPeerNetNameErr
				return
			}
			if uint(id) != uint(peerName) {
				config.Mapping[uint(peerName)] = peerNetName + ":" + port
				config.Peers = append(config.Peers, uint(peerName))
			}
		}
	}

	return
}

func invMain() {
	v, err := utils.InitConfig("INV", "/config")
	if err != nil {
		logrus.Errorf("Error parsing config file: %s", err)
		return
	}

	config, err := parseConfig(v)
	if err != nil {
		logrus.Errorf("Error parsing config file: %s", err)
	}
	invitation := invitation.Invitation(config)
	if err := invitation.Run(); err != nil {
		logrus.Fatalf("Invitation process ended with error: %s", err)
	}
}

func main() {
	client, _ := beater.NewBeaterClient("beater1", "127.0.0.1:9999")
	server := beater.NewBeaterServer([]string{"beater1"}, []string{"127.0.0.1"}, "10000")

	go client.Run()
	server.Run("9999")
}
