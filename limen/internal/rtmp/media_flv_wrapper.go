package rtmp

import (
	"bufio"
	"bytes"
)

type mediaFlvWrapper struct {
	headerInsterted bool
	audioPresent    bool
	videoPresent    bool
}

func NewMediaFlvWrapper(audioPresent bool, videoPresent bool) *mediaFlvWrapper {
	return &mediaFlvWrapper{headerInsterted: false, audioPresent: audioPresent, videoPresent: videoPresent}
}

func (w *mediaFlvWrapper) WrapMessage(message *Message) []byte {
	bytes := new(bytes.Buffer)

	writer := bufio.NewWriter(bytes)

	if !w.headerInsterted {
		w.insertHeader(writer)
		w.headerInsterted = true
	}

	w.insertBody(writer, message)

	writer.Flush()

	return bytes.Bytes()
}

func (w *mediaFlvWrapper) insertHeader(writer *bufio.Writer) {
	var flags uint8 = 0x00

	if w.audioPresent {
		flags |= 0x04
	}

	if w.videoPresent {
		flags |= 0x01
	}

	_, err := writer.Write([]byte{
		'F', 'L', 'V',
		// verstion
		0x01,
		// flags
		flags,
		// data offset
		0x0, 0x0, 0x0, 0x9,
		// PreviousTagSize
		0x0, 0x0, 0x0, 0x0,
	})
	if err != nil {
		panic("failed to write flv header")
	}
}

func (w *mediaFlvWrapper) insertBody(writer *bufio.Writer, message *Message) {
	header := message.Header

	tag_size := header.BodySize + 11
	metadata := []byte{
		header.Type,
		// body size:w
		byte(header.BodySize >> 16),
		byte(header.BodySize >> 8),
		byte(header.BodySize),
		// lower timestamp
		byte(header.Timestamp >> 16),
		byte(header.Timestamp >> 8),
		byte(header.Timestamp),
		// extended timestamp
		byte(header.Timestamp>>24) & 0xff,
		byte(header.StreamId>>16) & 0xff,
		byte(header.StreamId>>8) & 0xff,
		byte(header.StreamId) & 0xff,
	}

	if _, err := writer.Write(metadata); err != nil {
		panic("failed to write flv body metadata")
	}

	if _, err := writer.Write(message.Payload); err != nil {
		panic("failed to write flv body")
	}

	if _, err := writer.Write([]byte{
		byte(tag_size >> 24),
		byte(tag_size >> 16),
		byte(tag_size >> 8),
		byte(tag_size),
	}); err != nil {
		panic("failed to write flv tag size")
	}
}
