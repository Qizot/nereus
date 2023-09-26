package aac

type AACProfile uint8

const (
	ProfileMain    AACProfile = 0
	ProfileLC      AACProfile = 1
	ProfileSSR     AACProfile = 2
	ProfileLTP     AACProfile = 3
	ProfileHE      AACProfile = 4
	ProfileHEv2    AACProfile = 5
	ProfileUnknown AACProfile = 6
)

type Encapsulation uint8

const (
	EncapsulationNone Encapsulation = 0
	EncapsulationADTS Encapsulation = 1
)

type ConfigType uint8

const (
	ConfigTypeEsds          ConfigType = 0
	ConfigTypeAudioSpecific ConfigType = 1
)

type Format struct {
	SampleRate      uint32
	SamplesPerFrame uint32
	Profile         AACProfile
	Channels        uint8
	MpegVersion     uint8
	Encapsulation   Encapsulation
	Config          []byte
	ConfigType      ConfigType
}
