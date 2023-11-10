package utils

import "io"


type ChannelReader struct {
	Channel chan []byte
}

func NewChannelReader(from io.Reader) *ChannelReader {
	return nil
}