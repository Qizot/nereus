package aac

import (
	"errors"

	"limen/internal/util"
)

func ParseEsdsConfig(esds []byte) (*Format, error) {
	streamPriority := byte(0)

	reader := &util.BitReader{
		Data: esds,
	}

	section3, err := extractEsdsSection(reader, 3)
	if err != nil {
		return nil, err
	}

	if len(section3) < 3 || section3[2] != streamPriority {
		return nil, errors.New("invalid esds section 3")
	}

	section4, err := extractEsdsSection(reader, 4)
	if err != nil {
		return nil, err
	}

	section6, err := extractEsdsSection(reader, 6)
	if err != nil {
		return nil, err
	}

	if len(section6) != 1 || section6[0] != 0x02 {
		return nil, errors.New("invalid esds section 6")
	}

	// mpeg4 audio
	objectTypeId := byte(64)
	// audio type
	streamType := byte(5)
	upstreamFlag := byte(0)
	reservedFlag := byte(1)
	bufferSize := uint32(0)

	if len(section4) < 13 {
		return nil, errors.New("invalid esds section 4 size")
	}

	sectionReader := util.BitReader{
		Data: section4,
	}

	var readerPayload uint64

	sectionReader.ReadBits(8, &readerPayload)
	readObjectTypeId := byte(readerPayload)

	sectionReader.ReadBits(6, &readerPayload)
	readStreamType := byte(readerPayload)

	sectionReader.ReadBits(1, &readerPayload)
	readUpstreamFlag := byte(readerPayload)

	sectionReader.ReadBits(1, &readerPayload)
	readReservedFlag := byte(readerPayload)

	sectionReader.ReadBits(24, &readerPayload)
	readBufferSize := uint32(readerPayload)

	if readObjectTypeId != objectTypeId ||
		readStreamType != streamType ||
		readUpstreamFlag != upstreamFlag ||
		readReservedFlag != reservedFlag ||
		readBufferSize != bufferSize {
		return nil, errors.New("invalid esds section 4")
	}

	// skip max bitrate and avg bitrate
	reader.SkipBits(64)

	sectionFiveSize := reader.BitsAvailable() / 8
	section5Data := make([]byte, sectionFiveSize)

	if !reader.ReadSlice(section5Data[:]) {
		return nil, errors.New("invalid section5")
	}

	sectionReader = util.BitReader{
		Data: section5Data,
	}

	section5, err := extractEsdsSection(&sectionReader, 5)
	if err != nil {
		return nil, err
	}

	format, err := ParseAudioSpecificConfig(section5)
	if err != nil {
		return nil, err
	}

	format.ConfigType = ConfigTypeEsds

	return format, nil
}

func ParseAudioSpecificConfig(data []byte) (*Format, error) {
	if len(data) < 2 {
		return nil, errors.New("audio specific config is too short")
	}

	var payload uint64
	reader := util.BitReader{
		Data: data,
	}

	reader.ReadBits(5, &payload)
	profile := byte(payload)

	reader.ReadBits(4, &payload)
	frequencyId := byte(payload)

	customFrequencyLen := 0
	if frequencyId == 15 {
		customFrequencyLen = 24
	}

	var customFrequency uint32
	if customFrequencyLen > 0 {
		reader.ReadBits(customFrequencyLen, &payload)
		customFrequency = uint32(payload)
	}

	reader.ReadBits(4, &payload)
	channelConfigId := byte(payload)

	reader.ReadBits(1, &payload)
	frameLengthId := byte(payload)

	var sampleRate uint32
	if frequencyId == 15 {
		sampleRate = customFrequency
	}

	format := &Format{
		SampleRate:      sampleRate,
		SamplesPerFrame: frameLengthIdToSamplesPerFrame(frameLengthId),
		Profile:         aotToProfile(profile),
		Channels:        channelConfigId,
		MpegVersion:     4,
		Encapsulation:   EncapsulationNone,
		Config:          data,
		ConfigType:      ConfigTypeAudioSpecific,
	}

	return format, nil
}

func frameLengthIdToSamplesPerFrame(frameLengthId byte) uint32 {
	switch frameLengthId {
	case 0:
		return 1024
	case 1:
		return 960
	default:
		return 0
	}
}

func aotToProfile(aot byte) AACProfile {
	switch aot {
	case 1:
		return ProfileMain
	case 2:
		return ProfileLC
	case 3:
		return ProfileSSR
	case 4:
		return ProfileLTP
	case 5:
		return ProfileHE
	case 29:
		return ProfileHEv2
	default:
		return ProfileUnknown
	}
}

func extractEsdsSection(reader *util.BitReader, section uint8) ([]byte, error) {
	var readPayload uint64

	if reader.BitsAvailable() < 5*8 {
		return nil, errors.New("esds payload too short")
	}

	reader.ReadBits(8, &readPayload)

	sectionNum := uint8(readPayload)

	if sectionNum != section {
		return nil, errors.New("invalid esds section number")
	}

	var tag [3]byte
	if !reader.ReadSlice(tag[:]) || tag[0] != 0x80 || tag[1] != 0x80 || tag[2] != 0x80 {
		return nil, errors.New("invalid esds type tag")
	}

	reader.ReadBits(8, &readPayload)
	dataLen := byte(readPayload)

	if reader.BitsAvailable()*8 < int(dataLen) {
		return nil, errors.New("too small esds section")
	}

	payload := make([]byte, dataLen)

	reader.ReadSlice(payload)

	return payload, nil
}
