package rtmp

import (
	"bufio"
	"bytes"
)

type messageWriter struct {
	chunkSize int32
}

func NewMessageWriter() *messageWriter {
	return &messageWriter{chunkSize: -1}
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

	if err = writer.Flush(); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func (m *messageWriter) SetChunkSize(size int32) {
	m.chunkSize = size
}

func (m *messageWriter) writePayload(writer *bufio.Writer, message *Message) error {
	chunkedPayload := chunkPayload(message, int(m.chunkSize))
	_, err := writer.Write(chunkedPayload)
	return err
}

func (m *messageWriter) writeHeader(writer *bufio.Writer, message *Message) error {
	_, err := writer.Write(serializeHeader(message.Header))

	return err
}

// NOTE: we are only serializing to type0 header
func serializeHeader(header *Header) []byte {
	return []byte{
		header.ChunkStreamId & 0x3f,
		byte(header.Timestamp>>16) & 0xff,
		byte(header.Timestamp>>8) & 0xff,
		byte(header.Timestamp) & 0xff,
		byte(header.BodySize>>16) & 0xff,
		byte(header.BodySize>>8) & 0xff,
		byte(header.BodySize) & 0xff,
		header.Type,
		byte(header.StreamId>>24) & 0xff,
		byte(header.StreamId>>16) & 0xff,
		byte(header.StreamId>>8) & 0xff,
		byte(header.StreamId) & 0xff,
	}
}

func chunkPayload(msg *Message, chunkSize int) []byte {
	if chunkSize == -1 || len(msg.Payload) <= chunkSize {
		return msg.Payload
	}

	chunks := int((len(msg.Payload) - 1) / chunkSize)
	var extendedTimestamps int
	if msg.Header.ExtendedTimestamp {
		extendedTimestamps = chunks
	} else {
		extendedTimestamps = 0
	}

	newPayload := make([]byte, len(msg.Payload)+chunks+4*extendedTimestamps)

	copy(newPayload[:chunkSize], msg.Payload[:chunkSize])

	offset := chunkSize
	for i := 1; i <= chunks; i++ {
		// write the separator
		newPayload[offset] = 0b11000000 | msg.Header.ChunkStreamId
		offset += 1

		// write the timestamp
		if msg.Header.ExtendedTimestamp {
			newPayload[offset] = byte(msg.Header.Timestamp>>24) & 0xff
			newPayload[offset+1] = byte(msg.Header.Timestamp>>16) & 0xff
			newPayload[offset+2] = byte(msg.Header.Timestamp>>8) & 0xff
			newPayload[offset+3] = byte(msg.Header.Timestamp) & 0xff

			offset += 4
		}

		var offsetTo int
		if i*chunkSize+chunkSize > len(msg.Payload) {
			offsetTo = len(msg.Payload)
		} else {
			offsetTo = i*chunkSize + chunkSize
		}

		size := offsetTo - i*chunkSize

		copy(newPayload[offset:offset+size], msg.Payload[i*chunkSize:offsetTo])

		offset += size
	}

	return newPayload
}
