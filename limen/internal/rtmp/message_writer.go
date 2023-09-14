package rtmp

import (
	"bufio"
	"bytes"
)

type messageWriter struct {
	lastHeader *Header
}

func NewMessageWriter() *messageWriter {
	return &messageWriter{}
}

func (m *messageWriter) Write(message *Message) ([]byte, error) {
	var buffer bytes.Buffer
	writer := bufio.NewWriter(&buffer)
	err := m.writeHeader(writer, message)
	if err != nil {
		return nil, err
	}

	err = m.writePayload(writer, message)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func (m *messageWriter) writePayload(writer *bufio.Writer, message *Message) error {
	return nil

}

func (m *messageWriter) writeHeader(writer *bufio.Writer, message *Message) error {
	return nil
}
