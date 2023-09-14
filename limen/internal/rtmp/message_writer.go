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
	_, err := writer.Write(message.Payload)
	return err
}

func (m *messageWriter) writeHeader(writer *bufio.Writer, message *Message) error {
	_, err := writer.Write(serializeHeader(message.Header))

	return err
}

// NOTE: we are only serializing to type0 header
func serializeHeader(header *Header) []byte {
	// TODO: find a better way to do this...
	return []byte{
		header.ChunkStreamId & 0x3F,
		byte(header.Timestamp>>16) & 0xFF,
		byte(header.Timestamp>>8) & 0xFF,
		byte(header.Timestamp) & 0xFF,
		byte(header.BodySize>>16) & 0xFF,
		byte(header.BodySize>>8) & 0xFF,
		byte(header.BodySize) & 0xFF,
		header.Type,
		byte(header.StreamId>>24) & 0xFF,
		byte(header.StreamId>>16) & 0xFF,
		byte(header.StreamId>>8) & 0xFF,
		byte(header.StreamId) & 0xFF,
	}
}
