package rtmp

type Header struct {
	ChunkStreamId     uint8
	Timestamp         uint32
	TimestampDelta    uint32
	BodySize          uint32
	Length            uint32
	Type              uint8
	StreamId          uint32
	ExtendedTimestamp bool
}
