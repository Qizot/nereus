package rtmp

import (
	"encoding/binary"
)

type SetChunkSizeMessage struct {
	ChunkSize uint32
}

func (c *SetChunkSizeMessage) Type() uint8 {
	return SetChunkSizeType
}

func (c *SetChunkSizeMessage) Serialize() []byte {
	payload := make([]byte, 4)
	binary.BigEndian.PutUint32(payload, c.ChunkSize)

	payload[0] &= 0x7F

	return payload
}

func (c *SetChunkSizeMessage) Deserialize(data []byte) error {
	if len(data) != 4 {
		return ErrInvalidMessageFormat
	}

	c.ChunkSize = binary.BigEndian.Uint32(data)

	return nil
}
