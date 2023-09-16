package rtmp

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"io"
)

type handshake struct {
	c1 []byte
	s1 []byte
}

func NewHandshake() *handshake {
	return &handshake{}
}

func (h *handshake) ReceiveC0C1(buffer *bufio.Reader) error {
	var payload [1537]byte

	if _, err := io.ReadFull(buffer, payload[:]); err != nil {
		return ErrNotEnoughData
	}

	if payload[0] != 0x03 {
		return ErrInvalidHandshake
	}

	h.c1 = payload[1:]

	return nil
}

func (h *handshake) GenerateS0S1() []byte {
	buf := make([]byte, 1537)
	buf[0] = 0x03
	_, _ = rand.Read(buf[1:])

	h.s1 = buf[1:]

	return buf
}

func (h *handshake) ReceiveC2(buffer *bufio.Reader) error {
	var c2 [1536]byte

	if _, err := io.ReadFull(buffer, c2[:]); err != nil {
		return ErrNotEnoughData
	}

	if !bytes.Equal(c2[:], h.s1) {
		return ErrInvalidHandshake
	}

	return nil
}

func (h *handshake) GetS2() []byte {
	return h.c1
}
