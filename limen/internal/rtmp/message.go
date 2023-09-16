package rtmp

const (
	SetChunkSizeType     = 0x1
	UserControlType      = 0x4
	WindowAckSizeType    = 0x5
	SetPeerBandwidthType = 0x6
	AudioType            = 0x8
	VideoType            = 0x9
	AmfDataType          = 0x12
	AmfCommandType       = 0x14
)

type Message struct {
	Header  *Header
	Payload []byte
}

type MessageSerializer interface {
	Serialize() []byte
	Type() uint8
}
