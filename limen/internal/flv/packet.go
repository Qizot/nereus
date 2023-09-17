package flv

type (
	PacketType uint8
	SoundType  uint8
)

const (
	AudioPacket       PacketType = 0
	VideoPacket       PacketType = 1
	ScriptDataPacket  PacketType = 2
	AudioConfigPacket PacketType = 3
	VideoConfigPacket PacketType = 4
)

const (
	MonoSound   SoundType = 0
	StereoSound SoundType = 1
)

type Packet struct {
	CodecParams interface{}
	Data        []byte
	Pts         int
	Dts         int
	StreamId    uint32
	Type        PacketType
	Codec       uint8
}

type VideoCodecParams struct {
	KeyFrame        bool
	CompositionTime int
}

type AudioCodecParams struct {
	SoundRate uint32
	SoundType SoundType
}

func (p *Packet) SetTimestamps(timestamp int) {
	p.Dts = timestamp
	p.Pts = timestamp

	if p.Type == VideoPacket || p.Type == VideoConfigPacket {
		if params, ok := p.CodecParams.(*VideoCodecParams); ok {
			p.Pts += params.CompositionTime
		}
	}
}
