package flv

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

var (
	ErrNotEnoughData            = errors.New("not enough data")
	ErrEOF                      = errors.New("EOF")
	ErrMalformedPacket          = errors.New("malformed packet")
	ErrInvalidStreamId          = errors.New("invalid stream id")
	ErrUnsupportedScriptData    = errors.New("unsupported script data")
	ErrInvalidPacketPayloadType = errors.New("invalid packet payload type")
	ErrExtFormatUnsupported     = errors.New("extended flv format unsupported")
)

type decoder struct {
	headerSeen   bool
	audioPresent bool
	videoPresent bool
}

func NewFlvDecoder() *decoder {
	return &decoder{headerSeen: false, audioPresent: false, videoPresent: true}
}

func (d *decoder) Decode(buffer *bufio.Reader) (*Packet, error) {
	if !d.headerSeen {
		if err := d.decodeHeader(buffer); err != nil {
			return nil, err
		}

		d.headerSeen = true
	}

	return d.decodeBody(buffer)
}

func (d *decoder) decodeHeader(buffer *bufio.Reader) error {
	var header [9]byte

	_, err := io.ReadFull(buffer, header[:])

	// NOTE: we should be able to at least decode the header and one body
	if errors.Is(err, io.ErrUnexpectedEOF) {
		return ErrMalformedPacket
	}

	if !bytes.Equal(header[:3], []byte("FLV")) {
		return ErrMalformedPacket
	}

	flags := header[4]

	d.audioPresent = flags&0b00000100 > 0
	d.videoPresent = flags&0b00000001 > 0

	bytesToDiscard := int(binary.BigEndian.Uint32(header[5:9])) - 9

	_, err = buffer.Discard(bytesToDiscard)
	if err != nil {
		return ErrMalformedPacket
	}

	return nil
}

func (d *decoder) decodeBody(buffer *bufio.Reader) (*Packet, error) {
	// head: 4 + flags: 1 + datasize: 3 + timestamp: 3 +  timestamp_extended: 1 + streamid 3
	var bodyHeader [15]byte
	n, err := io.ReadFull(buffer, bodyHeader[:])
	if n == 0 {
		return nil, ErrNotEnoughData
	}

	if err != nil {
		return nil, ErrMalformedPacket
	}

	dataSize := int(decodeUint24(bodyHeader[5:8]))

	const scriptDataTag = 0b00010010

	if bodyHeader[4]&0x3f == scriptDataTag {
		n, err := buffer.Discard(dataSize)
		if n != dataSize {
			return nil, ErrMalformedPacket
		}
		if err != nil {
			return nil, ErrEOF
		}

		return d.decodeBody(buffer)
	}

	packetType := bodyHeader[4] & 0x1f
	var timestamp [4]byte
	copy(timestamp[1:], bodyHeader[8:11])
	timestamp[0] = bodyHeader[11]
	timestampValue := binary.BigEndian.Uint32(timestamp[:])
	streamId := decodeUint24(bodyHeader[12:15])

	if streamId != 0 {
		return nil, ErrInvalidStreamId
	}

	payload := make([]byte, dataSize)
	n, _ = io.ReadFull(buffer, payload[:])
	if n != dataSize {
		return nil, ErrMalformedPacket
	}

	resolvedPacketType := d.resolvedPacketType(packetType)
	if packet, err := d.decodePayload(resolvedPacketType, payload); err != nil {
		return nil, err
	} else {
		packet.StreamId = streamId
		packet.SetTimestamps(int(timestampValue))

		return packet, nil
	}
}

func (d *decoder) decodePayload(packetType PacketType, payload []byte) (*Packet, error) {
	switch packetType {
	case AudioPacket, AudioConfigPacket:
		return d.decodeAudioPacket(payload)
	case VideoPacket, VideoConfigPacket:
		return d.decodeVideoPacket(payload)
	case ScriptDataPacket:
		return nil, ErrUnsupportedScriptData
	}

	panic(fmt.Sprintf("invalid packet type %d", packetType))
}

func (d *decoder) decodeAudioPacket(payload []byte) (*Packet, error) {
	if len(payload) < 2 {
		return nil, ErrMalformedPacket
	}

	soundFormat := payload[0] >> 4
	soundRate := (payload[0] >> 2) & 0x03
	// soundSize := (payload[0] >> 1) & 0x01
	soundType := payload[0] & 0x01
	packetType := payload[1]

	var configSoundRate uint32

	if !ValidateSoundFormat(soundFormat) {
		return nil, ErrInvalidPacketPayloadType
	}

	switch soundRate {
	case 0:
		configSoundRate = 5500
	case 1:
		configSoundRate = 11000
	case 2:
		configSoundRate = 22050
	case 3:
		configSoundRate = 44000
	default:
		configSoundRate = 0
	}

	var configSoundType SoundType

	if soundType == 0 {
		configSoundType = MonoSound
	} else {
		configSoundType = StereoSound
	}

	var configPacketType PacketType
	if packetType == 0 {
		configPacketType = AudioConfigPacket
	} else {
		configPacketType = AudioPacket
	}

	packet := &Packet{
		Data:  payload[2:],
		Codec: soundFormat,
		Type:  configPacketType,
		CodecParams: &AudioCodecParams{
			SoundRate: configSoundRate,
			SoundType: configSoundType,
		},
	}

	return packet, nil
}

func (d *decoder) decodeVideoPacket(payload []byte) (*Packet, error) {
	if len(payload) < 6 {
		return nil, ErrMalformedPacket
	}

	frameType := payload[0] >> 4
	codec := payload[0] & 0x0f
	packetType := payload[1]
	compositionTime := decodeUint24(payload[2:5])

	if !ValidateVideoCodec(codec) {
		return nil, ErrInvalidPacketPayloadType
	}

	if frameType&0x08 > 0 {
		return nil, ErrExtFormatUnsupported
	}

	var configPacketType PacketType
	if packetType == 0 {
		configPacketType = VideoConfigPacket
	} else {
		configPacketType = VideoPacket
	}

	keyFrame := frameType == 1

	packet := &Packet{
		Data:  payload[5:],
		Codec: codec,
		Type:  configPacketType,
		CodecParams: &VideoCodecParams{
			KeyFrame:        keyFrame,
			CompositionTime: int(compositionTime),
		},
	}

	return packet, nil
}

func (d *decoder) resolvedPacketType(packetType uint8) PacketType {
	switch packetType {
	case 8:
		return AudioPacket
	case 9:
		return VideoPacket
	case 18:
		return ScriptDataPacket
	}

	panic(fmt.Sprintf("invalid packet type %d", packetType))
}

func decodeUint24(data []byte) uint32 {
	var size [4]byte
	_ = copy(size[1:], data[:])
	size[0] = 0
	return binary.BigEndian.Uint32(size[:])
}
