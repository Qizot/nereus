package rtmp

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadMessageWithHeaderType0(t *testing.T) {
	payload := []byte{
		// type
		0x01,
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
	}

	reader := NewMessageReader()

	buffer := bufio.NewReader(bytes.NewReader(payload))
	message, err := reader.ReadMessage(buffer)

	assert.Nil(t, err)
	assert.Equal(t, message.Header.ChunkStreamId, uint8(0x01))
	assert.Equal(t, message.Header.Timestamp, uint32(0x000fffff))
	assert.Equal(t, message.Header.BodySize, uint32(1))
	assert.Equal(t, message.Header.Type, uint8(2))
	assert.Equal(t, message.Header.StreamId, uint32(1))
	assert.Equal(t, message.Payload, []byte{0xff})
}

func TestReadMessageWithHeaderType0ExtendedTimestamp(t *testing.T) {
	payload := []byte{
		// type
		0x01,
		// extended timestamp marker
		0xff, 0xff, 0xff,
		// body size
		0x0, 0x0, 0x01,
		// type
		0x2,
		// stream id
		0x00, 0x0, 0x0, 0x1,
		// extended timestmap
		0x00, 0xAA, 0xAA, 0xAA,
		// payload
		0xff,
	}

	reader := NewMessageReader()

	buffer := bufio.NewReader(bytes.NewReader(payload))
	message, err := reader.ReadMessage(buffer)

	assert.Nil(t, err)
	assert.Equal(t, message.Header.ChunkStreamId, uint8(0x01))
	assert.Equal(t, message.Header.Timestamp, uint32(0x00aaaaaa))
	assert.Equal(t, message.Header.TimestampDelta, uint32(0))
	assert.Equal(t, message.Header.BodySize, uint32(1))
	assert.Equal(t, message.Header.Type, uint8(2))
	assert.Equal(t, message.Header.StreamId, uint32(1))
	assert.Equal(t, message.Payload, []byte{0xff})
}

func TestReadMessageWithHeaderType1(t *testing.T) {
	payload := []byte{
		// type
		0b01000001,
		// timestmap delta
		0x00, 0x00, 0xff,
		// body size
		0x0, 0x0, 0x01,
		// type
		0x2,
		// payload
		0xff,
	}

	previousHeader := &Header{
		ChunkStreamId: 1,
		Timestamp:     2137,
		BodySize:      1,
		Type:          2,
		StreamId:      66,
	}

	reader := NewMessageReader()
	reader.lastHeader = previousHeader

	buffer := bufio.NewReader(bytes.NewReader(payload))
	message, err := reader.ReadMessage(buffer)

	assert.Nil(t, err)
	assert.Equal(t, message.Header.ChunkStreamId, uint8(0x01))
	assert.Equal(t, message.Header.Timestamp, previousHeader.Timestamp+uint32(0xff))
	assert.Equal(t, message.Header.TimestampDelta, uint32(0xff))
	assert.Equal(t, message.Header.BodySize, uint32(1))
	assert.Equal(t, message.Header.Type, uint8(2))
	assert.Equal(t, message.Header.StreamId, previousHeader.StreamId)
	assert.Equal(t, message.Payload, []byte{0xff})
}

func TestReadMessageWithHeaderType1WithExtendedTimestamp(t *testing.T) {
	payload := []byte{
		// type
		0b01000001,
		// extended timestamp marker
		0xff, 0xff, 0xff,
		// body size
		0x0, 0x0, 0x01,
		// type
		0x2,
		// extended timestamp
		0xbb, 0x00, 0x00, 0x00,
		// payload
		0xff,
	}

	previousHeader := &Header{
		ChunkStreamId: 1,
		Timestamp:     0x00aaaaaa,
		BodySize:      1,
		Type:          2,
		StreamId:      66,
	}

	reader := NewMessageReader()
	reader.lastHeader = previousHeader

	buffer := bufio.NewReader(bytes.NewReader(payload))
	message, err := reader.ReadMessage(buffer)

	assert.Nil(t, err)
	assert.Equal(t, message.Header.ChunkStreamId, uint8(0x01))
	assert.Equal(t, message.Header.Timestamp, previousHeader.Timestamp+uint32(0xbb000000))
	assert.Equal(t, message.Header.TimestampDelta, uint32(0xbb000000))
	assert.Equal(t, message.Header.BodySize, uint32(1))
	assert.Equal(t, message.Header.Type, uint8(2))
	assert.Equal(t, message.Header.StreamId, previousHeader.StreamId)
	assert.Equal(t, message.Payload, []byte{0xff})
}

func TestReadMessageWithHeaderType2(t *testing.T) {
	payload := []byte{
		// type
		0b10000001,
		// timestmap delta
		0x00, 0x00, 0xff,
		// payload
		0xff,
	}

	previousHeader := &Header{
		ChunkStreamId: 1,
		Timestamp:     2137,
		BodySize:      1,
		Type:          2,
		StreamId:      1,
	}

	reader := NewMessageReader()
	reader.lastHeader = previousHeader

	buffer := bufio.NewReader(bytes.NewReader(payload))
	message, err := reader.ReadMessage(buffer)

	assert.Nil(t, err)
	assert.Equal(t, message.Header.ChunkStreamId, uint8(0x01))
	assert.Equal(t, message.Header.Timestamp, previousHeader.Timestamp+uint32(0xff))
	assert.Equal(t, message.Header.TimestampDelta, uint32(0xff))
	assert.Equal(t, message.Header.BodySize, previousHeader.BodySize)
	assert.Equal(t, message.Header.Type, previousHeader.Type)
	assert.Equal(t, message.Header.StreamId, previousHeader.StreamId)
	assert.Equal(t, message.Payload, []byte{0xff})
}

func TestReadMessageWithHeaderType2WithExtendedTimestmap(t *testing.T) {
	payload := []byte{
		// type
		0b10000001,
		// extended timestamp marker
		0xff, 0xff, 0xff,
		// extended timestamp
		0xbb, 0x00, 0x00, 0x00,
		// payload
		0xff,
	}

	previousHeader := &Header{
		ChunkStreamId: 1,
		Timestamp:     0x00aaaaaa,
		BodySize:      1,
		Type:          2,
		StreamId:      1,
	}

	reader := NewMessageReader()
	reader.lastHeader = previousHeader

	buffer := bufio.NewReader(bytes.NewReader(payload))
	message, err := reader.ReadMessage(buffer)

	assert.Nil(t, err)
	assert.Equal(t, message.Header.ChunkStreamId, uint8(0x01))
	assert.Equal(t, message.Header.Timestamp, previousHeader.Timestamp+uint32(0xbb000000))
	assert.Equal(t, message.Header.TimestampDelta, uint32(0xbb000000))
	assert.Equal(t, message.Header.BodySize, previousHeader.BodySize)
	assert.Equal(t, message.Header.Type, previousHeader.Type)
	assert.Equal(t, message.Header.StreamId, previousHeader.StreamId)
	assert.Equal(t, message.Payload, []byte{0xff})
}

func TestReadMessageWithHeaderType3(t *testing.T) {
	payload := []byte{
		// type
		0b11000001,
		// payload
		0xff,
	}

	previousHeader := &Header{
		ChunkStreamId:  1,
		Timestamp:      2137,
		TimestampDelta: 1,
		BodySize:       1,
		Type:           2,
		StreamId:       1,
	}

	reader := NewMessageReader()
	reader.lastHeader = previousHeader

	buffer := bufio.NewReader(bytes.NewReader(payload))
	message, err := reader.ReadMessage(buffer)

	assert.Nil(t, err)
	assert.Equal(t, message.Header.ChunkStreamId, uint8(0x01))
	assert.Equal(t, message.Header.Timestamp, previousHeader.Timestamp+previousHeader.TimestampDelta)
	assert.Equal(t, message.Header.TimestampDelta, previousHeader.TimestampDelta)
	assert.Equal(t, message.Header.BodySize, previousHeader.BodySize)
	assert.Equal(t, message.Header.Type, previousHeader.Type)
	assert.Equal(t, message.Header.StreamId, previousHeader.StreamId)
	assert.Equal(t, message.Payload, []byte{0xff})
}

func TestReadMessageWithHeaderType3WithExtendedTimestamp(t *testing.T) {
	payload := []byte{
		// type
		0b11000001,
		// extended timestmap
		0xbb, 0x00, 0x00, 0x00,
		// payload
		0xff,
	}

	previousHeader := &Header{
		ChunkStreamId:     1,
		Timestamp:         0x00aaaaaa,
		ExtendedTimestamp: true,
		BodySize:          1,
		Type:              2,
		StreamId:          1,
	}

	reader := NewMessageReader()
	reader.lastHeader = previousHeader

	buffer := bufio.NewReader(bytes.NewReader(payload))
	message, err := reader.ReadMessage(buffer)

	assert.Nil(t, err)
	assert.Equal(t, message.Header.ChunkStreamId, uint8(0x01))
	assert.Equal(t, message.Header.Timestamp, previousHeader.Timestamp+uint32(0xbb000000))
	assert.Equal(t, message.Header.TimestampDelta, uint32(0xbb000000))
	assert.Equal(t, message.Header.BodySize, previousHeader.BodySize)
	assert.Equal(t, message.Header.Type, previousHeader.Type)
	assert.Equal(t, message.Header.StreamId, previousHeader.StreamId)
	assert.Equal(t, message.Payload, []byte{0xff})
}

func TestReturnErrorOnNotEnoughData(t *testing.T) {
	payload := []byte{
		// type
		0x01,
		// timestmap
		0x0f, 0xff, 0xff,
		// body size
		0x0, 0x0, 0x01,
		// type
		0x2,
		// missing stream id
	}

	reader := NewMessageReader()
	buffer := bufio.NewReader(bytes.NewReader(payload))
	_, err := reader.ReadMessage(buffer)
	assert.Equal(t, err, ErrNotEnoughData)
}

func TestReturnErrorOnNotEnoughDataExtendedTimestamp(t *testing.T) {
	payload := []byte{
		// type
		0x01,
		// extended timestamp marker
		0xff, 0xff, 0xff,
		// body size
		0x0, 0x0, 0x01,
		// type
		0x2,
		// stream id
		0x00, 0x0, 0x0, 0x1,
		// partial extended timestmap
		0x00, 0xAA,
		// missing payload
	}

	reader := NewMessageReader()
	buffer := bufio.NewReader(bytes.NewReader(payload))
	_, err := reader.ReadMessage(buffer)
	assert.Equal(t, err, ErrNotEnoughData)
}

func TestReadMessageWithHeaderType0AndChunkSize(t *testing.T) {
	payload := []byte{
		// type
		0x01,
		// timestmap
		0x0f, 0xff, 0xff,
		// body size
		0x0, 0x0, 0x03,
		// type
		0x2,
		// stream id
		0x00, 0x0, 0x0, 0x1,
		// payload
		0xff,
		0xff,
		0xc1,
		0xff,
	}

	reader := NewMessageReader()
	reader.SetChunkSize(2)

	buffer := bufio.NewReader(bytes.NewReader(payload))
	message, err := reader.ReadMessage(buffer)

	assert.Nil(t, err)
	assert.Equal(t, message.Header.ChunkStreamId, uint8(0x01))
	assert.Equal(t, message.Header.Timestamp, uint32(0x000fffff))
	assert.Equal(t, message.Header.BodySize, uint32(3))
	assert.Equal(t, message.Header.Type, uint8(2))
	assert.Equal(t, message.Header.StreamId, uint32(1))
	assert.Equal(t, message.Payload, []byte{0xff, 0xff, 0xff})
}

func TestReadMessageWithHeaderType0AndChunkSizeAndExtendedTimestamp(t *testing.T) {
	payload := []byte{
		// type
		0x01,
		// timestmap
		0xff, 0xff, 0xff,
		// body size
		0x0, 0x0, 0x03,
		// type
		0x2,
		// stream id
		0x00, 0x0, 0x0, 0x1,
		// extended timestamp
		0x0f, 0xff, 0xff, 0xff,
		// payload
		0xff,
		0xff,
		// marker
		0xc1,
		// extended timestamp
		0x0f, 0xff, 0xff, 0xff,
		0xff,
	}

	reader := NewMessageReader()
	reader.SetChunkSize(2)

	buffer := bufio.NewReader(bytes.NewReader(payload))
	message, err := reader.ReadMessage(buffer)

	assert.Nil(t, err)
	assert.Equal(t, message.Header.ChunkStreamId, uint8(0x01))
	assert.Equal(t, message.Header.Timestamp, uint32(0x0fffffff))
	assert.Equal(t, message.Header.BodySize, uint32(3))
	assert.Equal(t, message.Header.Type, uint8(2))
	assert.Equal(t, message.Header.StreamId, uint32(1))
	assert.Equal(t, message.Payload, []byte{0xff, 0xff, 0xff})
}
