package main

import (
	"invitation/invitation"
	"log"
	"net"
	"time"

	"github.com/sirupsen/logrus"
)

func ping_pong(at, peer, message string) error {
	addr, err := net.ResolveUDPAddr("udp", at)

	if err != nil {
		return err
	}

	peerAddr, err := net.ResolveUDPAddr("udp", peer)

	if err != nil {
		return err
	}

	sckt, err := net.ListenUDP("udp", addr)

	if err != nil {
		return err
	}
	sleepTime, _ := time.ParseDuration("300ms")
	for {
		written, err := sckt.WriteToUDP([]byte(message), peerAddr)
		log.Printf("Written: %d", written)
		if err != nil {
			return err
		}
		pong := make([]byte, 4)
		_, _, _ = sckt.ReadFromUDP(pong)
		log.Printf("Peer says: %s", string(pong))
		time.Sleep(sleepTime)
	}

}

func main1() {

	peers := []uint{2, 3, 4}
	peerMapping := map[uint]string{
		2: "127.0.0.1:10000",
		3: "127.0.0.1:10001",
		4: "127.0.0.1:10002",
	}
	id := 1
	addr, _ := net.ResolveUDPAddr("udp", "0.0.0.0:9999")
	conn, _ := net.ListenUDP("udp", addr)
	status := invitation.Invitation(peers, uint(id), peerMapping, conn)

	if err := status.Run(); err != nil {
		logrus.Fatalf("Error occured during run: %s", err)
	}
}

func main2() {

	peers := []uint{1, 3, 4}
	peerMapping := map[uint]string{
		1: "127.0.0.1:9999",
		3: "127.0.0.1:10001",
		4: "127.0.0.1:10002",
	}
	id := 2
	addr, _ := net.ResolveUDPAddr("udp", "0.0.0.0:10000")
	conn, _ := net.ListenUDP("udp", addr)
	status := invitation.Invitation(peers, uint(id), peerMapping, conn)

	if err := status.Run(); err != nil {
		logrus.Fatalf("Error occured during run: %s", err)
	}
}

func main3() {

	peers := []uint{4, 2, 1}
	peerMapping := map[uint]string{
		1: "127.0.0.1:9999",
		2: "127.0.0.1:10000",
		4: "127.0.0.1:10002",
	}
	id := 3
	addr, _ := net.ResolveUDPAddr("udp", "0.0.0.0:10001")
	conn, _ := net.ListenUDP("udp", addr)
	status := invitation.Invitation(peers, uint(id), peerMapping, conn)

	if err := status.Run(); err != nil {
		logrus.Fatalf("Error occured during run: %s", err)
	}
}

func main4() {

	peers := []uint{1, 2, 3}
	peerMapping := map[uint]string{
		2: "127.0.0.1:10000",
		3: "127.0.0.1:10001",
		1: "127.0.0.1:9999",
	}
	id := 4
	addr, _ := net.ResolveUDPAddr("udp", "0.0.0.0:10002")
	conn, _ := net.ListenUDP("udp", addr)
	status := invitation.Invitation(peers, uint(id), peerMapping, conn)

	if err := status.Run(); err != nil {
		logrus.Fatalf("Error occured during run: %s", err)
	}
}

func main() {
	logrus.Info("Starting")
	main4()
}
