package rtmp

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
)

const ExtendedTimestampMarker = 0xFFFFFF

type messageReader struct {
	// FIXME: this should be done on chunk basis instead of
	// being a global state
	lastHeader           *Header
	partialHeader        *Header
	readingHeader        bool
	currentPayloadBuffer []byte
	chunkSize            int32
}

func NewMessageReader() *messageReader {
	return &messageReader{readingHeader: true, chunkSize: -1}
}

func (r *messageReader) ReadMessage(buffer *bufio.Reader) (*Message, error) {
	// Read the header

	if r.partialHeader != nil {
		header, err := r.readPartialHeader(buffer)
		if err != nil {
			return nil, err
		}

		r.lastHeader = header
		r.currentPayloadBuffer = make([]byte, chunkedBodySize(header, r.chunkSize))
	} else if r.readingHeader {
		header, err := r.readHeader(buffer)
		if err != nil {
			return nil, err
		}

		r.lastHeader = header
		r.currentPayloadBuffer = make([]byte, chunkedBodySize(header, r.chunkSize))
	}

	r.readingHeader = false

	// Read the payload
	_, err := io.ReadFull(buffer, r.currentPayloadBuffer)
	if err != nil {
		return nil, ErrNotEnoughData
	}

	r.currentPayloadBuffer = unchunkPayload(r.lastHeader, r.currentPayloadBuffer, r.chunkSize)

	// Payload has been read, switch to reading header
	r.readingHeader = true

	payload := r.currentPayloadBuffer
	r.currentPayloadBuffer = nil

	// Return the message
	return &Message{
		Header:  r.lastHeader,
		Payload: payload,
	}, nil
}

func (r *messageReader) SetChunkSize(size int32) {
	r.chunkSize = size
}

func (r *messageReader) readHeader(buffer *bufio.Reader) (*Header, error) {
	first_byte, err := buffer.ReadByte()
	if err != nil {
		return nil, ErrNotEnoughData
	}

	headerType := (first_byte & 0b11000000) >> 6
	chunkStreamId := first_byte & 0b00111111

	switch headerType {
	case 0:
		return r.readHeaderType0(buffer, chunkStreamId)
	case 1:
		return r.readHeaderType1(buffer, chunkStreamId)
	case 2:
		return r.readHeaderType2(buffer, chunkStreamId)
	case 3:
		return r.readHeaderType3(buffer, chunkStreamId)
	default:
		return nil, ErrInvalidHeaderType
	}
}

func (r *messageReader) readHeaderType0(buffer *bufio.Reader, chunkStreamId byte) (*Header, error) {
	var buff [11]byte
	_, err := io.ReadFull(buffer, buff[:])
	if err != nil {
		return nil, ErrNotEnoughData
	}

	header := &Header{
		ChunkStreamId: chunkStreamId,
		Timestamp:     binary.BigEndian.Uint32(buff[0:4]) >> 8,
		BodySize:      binary.BigEndian.Uint32(buff[3:7]) >> 8,
		Type:          buff[6],
		StreamId:      binary.BigEndian.Uint32(buff[7:11]),
	}

	if header.Timestamp == ExtendedTimestampMarker {
		header.ExtendedTimestamp = true

		extendedTimestamp, err := r.readExtendedTimestamp(buffer, header)
		if err != nil {
			return nil, err
		}

		header.Timestamp = extendedTimestamp
	}

	return header, nil
}

func (r *messageReader) readHeaderType1(buffer *bufio.Reader, chunkStreamId byte) (*Header, error) {
	if r.lastHeader == nil {
		return nil, ErrOtherHeaderTypeExpected
	}

	var buff [7]byte
	_, err := io.ReadFull(buffer, buff[:])
	if err != nil {
		return nil, ErrNotEnoughData
	}

	timestampDelta := (binary.BigEndian.Uint32(buff[0:4]) >> 8)
	header := &Header{
		ChunkStreamId:  chunkStreamId,
		Timestamp:      r.lastHeader.Timestamp + timestampDelta,
		TimestampDelta: timestampDelta,
		BodySize:       binary.BigEndian.Uint32(buff[3:7]) >> 8,
		Type:           buff[6],
		StreamId:       r.lastHeader.StreamId,
	}

	if timestampDelta == ExtendedTimestampMarker {
		header.ExtendedTimestamp = true

		timestampDelta, err := r.readExtendedTimestamp(buffer, header)
		if err != nil {
			return nil, err
		}

		header.Timestamp = r.lastHeader.Timestamp + timestampDelta
		header.TimestampDelta = timestampDelta
	}

	return header, nil
}

func (r *messageReader) readHeaderType2(buffer *bufio.Reader, chunkStreamId byte) (*Header, error) {
	if r.lastHeader == nil {
		return nil, ErrOtherHeaderTypeExpected
	}

	var buff [3]byte
	_, err := io.ReadFull(buffer, buff[:])
	if err != nil {
		return nil, ErrNotEnoughData
	}

	timestampDelta := (uint32(buff[0])<<16 | uint32(buff[1])<<8 | uint32(buff[2]))
	header := &Header{
		ChunkStreamId:  chunkStreamId,
		Timestamp:      r.lastHeader.Timestamp + timestampDelta,
		TimestampDelta: timestampDelta,
		BodySize:       r.lastHeader.BodySize,
		Type:           r.lastHeader.Type,
		StreamId:       r.lastHeader.StreamId,
	}

	if timestampDelta == ExtendedTimestampMarker {
		header.ExtendedTimestamp = true

		timestampDelta, err := r.readExtendedTimestamp(buffer, header)
		if err != nil {
			return nil, err
		}

		header.Timestamp = r.lastHeader.Timestamp + timestampDelta
		header.TimestampDelta = timestampDelta
	}

	return header, nil
}

func (r *messageReader) readHeaderType3(buffer *bufio.Reader, chunkStreamId byte) (*Header, error) {
	if r.lastHeader == nil {
		return nil, ErrOtherHeaderTypeExpected
	}

	header := &Header{
		ChunkStreamId:  chunkStreamId,
		Timestamp:      r.lastHeader.Timestamp + r.lastHeader.TimestampDelta,
		TimestampDelta: r.lastHeader.TimestampDelta,
		BodySize:       r.lastHeader.BodySize,
		Type:           r.lastHeader.Type,
		StreamId:       r.lastHeader.StreamId,
	}

	if r.lastHeader.ExtendedTimestamp {
		timestampDelta, err := r.readExtendedTimestamp(buffer, header)
		if err != nil {
			return nil, err
		}

		header.Timestamp = r.lastHeader.Timestamp + timestampDelta
		header.TimestampDelta = timestampDelta
	}

	return header, nil
}

func (r *messageReader) readPartialHeader(buffer *bufio.Reader) (*Header, error) {
	header := r.partialHeader
	extendedTimestamp, err := r.readExtendedTimestamp(buffer, header)
	if err != nil {
		return nil, err
	}

	r.partialHeader = nil

	switch header.Type {
	case 0:
		header.Timestamp = extendedTimestamp
	case 1, 2, 3:
		header.Timestamp = r.lastHeader.Timestamp + extendedTimestamp
		header.TimestampDelta = extendedTimestamp
	}

	return header, nil
}

func (r *messageReader) readExtendedTimestamp(buffer *bufio.Reader, header *Header) (uint32, error) {
	var extendedTimestamp [4]byte

	_, err := io.ReadFull(buffer, extendedTimestamp[:])
	if err != nil {
		r.partialHeader = header
		return 0, ErrNotEnoughData
	}

	return binary.BigEndian.Uint32(extendedTimestamp[:]), nil
}

func chunkedBodySize(header *Header, chunkSize int32) uint32 {
	if chunkSize == -1 {
		return header.BodySize
	}

	chunks := uint32(int32(header.BodySize-1) / chunkSize)

	var extendedTimestamps uint32 = 0
	if header.ExtendedTimestamp {
		extendedTimestamps = 4 * chunks
	}

	return header.BodySize + chunks + extendedTimestamps
}

func unchunkPayload(header *Header, chunked []byte, chunkSize int32) []byte {
	if chunkSize == -1 {
		return chunked
	}

	payload := make([]byte, header.BodySize)

	left := int(header.BodySize)
	offset := 0
	chunkedOffset := 0
	for left > 0 {
		size := int(chunkSize)

		if size > left {
			size = left
		}

		copy(payload[offset:offset+size], chunked[chunkedOffset:chunkedOffset+size])

		offset += size

		chunkedOffset += 1
		if header.ExtendedTimestamp {
			chunkedOffset += 4
		}
		chunkedOffset += size

		left -= size
	}

	return payload
}
