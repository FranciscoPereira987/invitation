package utils

import (
	"encoding/binary"
	"net"
)

func getLength(stream []byte) int {
	if len(stream) != 4 {
		return 0
	}
	return int(binary.LittleEndian.Uint32(stream))
}

func addLength(stream []byte) []byte {
	length := binary.LittleEndian.AppendUint32(nil, uint32(len(stream)))
	return append(length, stream...)
}

func SafeWriteTo(stream []byte, sckt *net.UDPConn, to *net.UDPAddr) error {

	n, err := sckt.WriteTo(addLength(stream), to)
	if n != len(stream) {
		return err
	}
	return nil
}

func readBuf(buf []byte, sckt *net.UDPConn) (*net.UDPAddr, error) {
	n, addr, err := sckt.ReadFromUDP(buf)

	if n == 0 {
		return nil, err
	}
	if n >= 4 {
		return addr, nil
	}
	return addr, err
}

func SafeReadFrom(sckt *net.UDPConn) ([]byte, *net.UDPAddr, error) {
	buf := make([]byte, 1024)
	from, err := readBuf(buf, sckt)

	if err == nil {
		length := getLength(buf[:4])
		buf = buf[4 : 4+length]
	}

	return buf, from, err
}
