package utils

import "net"

func Merge(c ...<-chan *net.UDPAddr) chan *net.UDPAddr {

	out := make(chan *net.UDPAddr)
	for _, channel := range c {
		go func(from <-chan *net.UDPAddr) {
			for msg := range from {
				out <- msg
			}
		}(channel)
	}
	return out
}
