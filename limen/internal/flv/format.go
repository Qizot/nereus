package flv

const (
	SoundFormatPCM              uint8 = 0
	SoundFormatADPCM            uint8 = 1
	SoundFormatMP3              uint8 = 2
	SoundFormatPCMLE            uint8 = 3
	SoundFormatNellymoser16Mono uint8 = 4
	SoundFormatNellymoser8Mono  uint8 = 5
	SoundTypeNellymoser         uint8 = 6
	SoundTypeG711ALaw           uint8 = 7
	SoundTypeG711MULaw          uint8 = 8
	SoundTypeAAC                uint8 = 10
	SoundTypeSpeex              uint8 = 11
	SoundTypeMP38k              uint8 = 14
	SoundTypeDeviceSpecific     uint8 = 15
)

type AudioSoundFormat uint8

func ValidateSoundFormat(soundFormat uint8) bool {
	return soundFormat <= 15 && soundFormat != 9 && soundFormat != 12 && soundFormat != 13
}

const (
	VideoCodecSOresonH263  uint8 = 2
	VideoCodecScreenVideo  uint8 = 3
	VideoCodecVP6          uint8 = 4
	VideoCodecVP6WithAlpha uint8 = 5
	VideoCodecScreenVideo2 uint8 = 6
	VideoCodecH264         uint8 = 7
)

type VideoCodec = uint8

func ExtVideoCodecAV1() [4]byte  { return [4]byte{'a', 'v', '0', '1'} }
func ExtVideoCodecVP9() [4]byte  { return [4]byte{'v', 'p', '0', '9'} }
func ExtVideoCodecHEVC() [4]byte { return [4]byte{'h', 'v', 'c', '1'} }

type ExtVideoCodec = [4]byte

func ValidateVideoCodec(videoCodec uint8) bool {
	return videoCodec <= 7 && videoCodec != 1
}
