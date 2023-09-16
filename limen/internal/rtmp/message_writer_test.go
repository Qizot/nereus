package rtmp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteMessage(t *testing.T) {
	msg := &Message{
		Header: &Header{
			Type:      0x2,
			Timestamp: 0x0fffff,
			BodySize:  1,
			StreamId:  0x00000001,
		},
		Payload: []byte{0xff},
	}

	writer := NewMessageWriter()

	payload, err := writer.Write(msg)
	assert.NotNil(t, payload)
	assert.Nil(t, err)

	assert.Equal(t, payload, []byte{
		// header type
		0x00,
		// timestmap
		0x0f, 0xff, 0xff,
		// body size
		0x0, 0x0, 0x01,
		// type
		0x2,
		// stream id
		0x00, 0x0, 0x0, 0x1,
		// payload
		0xff,
	})
}

func TestWriteMessageWithChunkSize(t *testing.T) {
	msg := &Message{
		Header: &Header{
			Type:          0x2,
			Timestamp:     0x0fffff,
			BodySize:      3,
			StreamId:      0x00000001,
			ChunkStreamId: 6,
		},
		Payload: []byte{0xff, 0xff, 0xff},
	}

	writer := NewMessageWriter()

	writer.SetChunkSize(2)

	payload, err := writer.Write(msg)
	assert.NotNil(t, payload)
	assert.Nil(t, err)

	assert.Equal(t, payload, []byte{
		// header type
		0x0 | (msg.Header.ChunkStreamId & 0x3f),
		// timestmap
		0x0f, 0xff, 0xff,
		// body size
		0x0, 0x0, 0x03,
		// type
		0x2,
		// stream id
		0x00, 0x0, 0x0, 0x1,
		// payload start
		// payload
		0xff,
		0xff,
		// marker
		0b11000000 | msg.Header.ChunkStreamId,
		// payload
		0xff,
	})
}

func TestWriteMessageWithChunkSizeAndExtendedTimestamp(t *testing.T) {
	msg := &Message{
		Header: &Header{
			Type:              0x2,
			Timestamp:         0x0fffff,
			BodySize:          3,
			StreamId:          0x00000001,
			ChunkStreamId:     6,
			ExtendedTimestamp: true,
		},
		Payload: []byte{0xff, 0xff, 0xff},
	}

	writer := NewMessageWriter()

	writer.SetChunkSize(2)

	payload, err := writer.Write(msg)
	assert.NotNil(t, payload)
	assert.Nil(t, err)

	assert.Equal(t, payload, []byte{
		// header type
		0x0 | (msg.Header.ChunkStreamId & 0x3f),
		// timestmap
		0x0f, 0xff, 0xff,
		// body size
		0x0, 0x0, 0x03,
		// type
		0x2,
		// stream id
		0x00, 0x0, 0x0, 0x1,
		// payload start
		// payload
		0xff,
		0xff,
		// marker
		0b11000000 | msg.Header.ChunkStreamId,
		// extended timestamp
		0x00, 0x0f, 0xff, 0xff,
		// payload
		0xff,
	})
}
