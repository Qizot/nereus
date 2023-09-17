package flv

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	HeaderAudioFlag = 0b00000100
	HeaderVideoFlag = 0b00000001
)

func HeaderFixture() []byte {
	return []byte{
		// header
		'F', 'L', 'V',
		0x01,
		// flags
		HeaderAudioFlag & HeaderVideoFlag,
		// data offset
		0x0, 0x0, 0x0, 0x09,
	}
}

func TestDecodeVideoHeader(t *testing.T) {
	header := []byte{
		// header
		'F', 'L', 'V',
		0x01,
		// flags
		HeaderAudioFlag,
		// data offset
		0x0, 0x0, 0x0, 0x09,
	}

	decoder := NewFlvDecoder()

	reader := bufio.NewReader(bytes.NewBuffer(header))
	packet, err := decoder.Decode(reader)

	assert.Nil(t, packet)
	assert.Equal(t, err, ErrNotEnoughData)
	assert.Equal(t, decoder.audioPresent, true)
	assert.Equal(t, decoder.videoPresent, false)
}

func TestDecodeSkippingScriptData(t *testing.T) {
	header := HeaderFixture()

	decoder := NewFlvDecoder()
	reader := bufio.NewReader(bytes.NewBuffer(header))
	_, err := decoder.Decode(reader)
	assert.Equal(t, err, ErrNotEnoughData)

	payload := []byte{
		// head
		0x0, 0x0, 0x0, 0x0,
		// scripting data flag
		0b00010010,
		// data size
		0x0, 0x0, 0x3,
		// timestamp
		0x0, 0x0, 0x1,
		// timestamp extended
		0x0,
		// stream id
		0x0, 0x0, 0x0,
		// payload
		0xB, 0xA, 0xD,
	}

	reader.Reset(bytes.NewBuffer(payload))

	packet, err := decoder.Decode(reader)
	assert.Nil(t, packet)
	assert.Equal(t, ErrNotEnoughData, err)
}

func TestDecodeAudioPacket(t *testing.T) {
	header := HeaderFixture()

	decoder := NewFlvDecoder()
	reader := bufio.NewReader(bytes.NewBuffer(header))
	_, err := decoder.Decode(reader)
	assert.Equal(t, err, ErrNotEnoughData)

	payload := []byte{
		// head
		0x0, 0x0, 0x0, 0x0,
		// flags
		0x8,
		// data size
		0x0, 0x0, 0x4,
		// timestamp
		0x0, 0x0, 0x1,
		// timestamp extended
		0x0,
		// stream id
		0x0, 0x0, 0x0,
		// payload
		// sound format: 4, sound rate: 2, sound size: 1, sound type: 1
		(SoundTypeAAC << 4) | 0b000001100 | 0b00000010 | 0b00000001,
		// packet type
		0x1,
		// payload
		0xff, 0xff,
	}

	reader.Reset(bytes.NewBuffer(payload))
	packet, err := decoder.Decode(reader)
	assert.Nil(t, err)
	assert.Equal(t, 1, packet.Dts)
	assert.Equal(t, 1, packet.Pts)
	assert.Equal(t, []byte{0xff, 0xff}, packet.Data)
	assert.Equal(t, SoundTypeAAC, packet.Codec)
	assert.Equal(t, AudioPacket, packet.Type)
	assert.Equal(t, uint32(44000), packet.CodecParams.(*AudioCodecParams).SoundRate)
}

func TestDecodeVideoPacket(t *testing.T) {
	header := HeaderFixture()

	decoder := NewFlvDecoder()
	reader := bufio.NewReader(bytes.NewBuffer(header))
	_, err := decoder.Decode(reader)
	assert.Equal(t, err, ErrNotEnoughData)

	payload := []byte{
		// head
		0x0, 0x0, 0x0, 0x0,
		// flags
		0x9,
		// data size
		0x0, 0x0, 0x7,
		// timestamp
		0x0, 0x0, 0x1,
		// timestamp extended
		0x0,
		// stream id
		0x0, 0x0, 0x0,
		// payload
		// frame type: 4, codec type: 4
		(1 << 4) | (VideoCodecH264),
		// packet type
		0x1,
		// composite time
		0x0, 0x0, 0x1,
		// payload
		0xff, 0xff,
	}

	reader.Reset(bytes.NewBuffer(payload))
	packet, err := decoder.Decode(reader)
	assert.Nil(t, err)
	assert.Equal(t, 1, packet.Dts)
	assert.Equal(t, 2, packet.Pts)
	assert.Equal(t, []byte{0xff, 0xff}, packet.Data)
	assert.Equal(t, VideoCodecH264, packet.Codec)
	assert.True(t, packet.CodecParams.(*VideoCodecParams).KeyFrame)
	assert.Equal(t, 1, packet.CodecParams.(*VideoCodecParams).CompositionTime)
}

func TestDecodeExtVideoPacket(t *testing.T) {
	header := HeaderFixture()

	decoder := NewFlvDecoder()
	reader := bufio.NewReader(bytes.NewBuffer(header))
	_, err := decoder.Decode(reader)
	assert.Equal(t, err, ErrNotEnoughData)

	payload := []byte{
		// head
		0x0, 0x0, 0x0, 0x0,
		// flags
		0x9,
		// data size
		0x0, 0x0, 0x7,
		// timestamp
		0x0, 0x0, 0x1,
		// timestamp extended
		0x0,
		// stream id
		0x0, 0x0, 0x0,
		// payload
		// frame type: 4, codec type: 4
		0x80 | (1 << 4) | (VideoCodecH264 & 0x0f),
		// packet type
		0x1,
		// composite time
		0x0, 0x0, 0x1,
		// payload
		0xff, 0xff,
	}

	reader.Reset(bytes.NewBuffer(payload))
	_, err = decoder.Decode(reader)
	assert.Equal(t, ErrExtFormatUnsupported, err)
}
