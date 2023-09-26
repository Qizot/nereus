package codec

type CodecType uint8

type FrameType uint8

type Frame struct {
	Metadata interface{}
	Data     []byte
	Dts      int
	Pts      int
	Codec    CodecType
	Type     FrameType
}
